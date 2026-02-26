package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

func Start() {
	sdk := k8s.NewK8sClient()
	metricResolution := facade.Config.GetDuration("metrics.resolution")
	metriceDuration := facade.Config.GetDuration("metrics.duration") //定时抓取
	controller, err := NewMetricsControlller(sdk.Sdk, metricResolution, metriceDuration)
	if err != nil {
		slog.Error("failed to create metrics controller", "err", err)
		return
	}

	stopCh := make(chan struct{})
	err = controller.Start(stopCh)
	if err != nil {
		slog.Error("failed to start metrics controller", "err", err)
	}
}

type MetricsControlller struct {
	sdk              *k8s.Sdk
	nodeInformer     cache.SharedIndexInformer
	nodeLister       v1lister.NodeLister
	informerFactory  informers.SharedInformerFactory
	metricResolution time.Duration
	metriceDuration  time.Duration
	tickStatusMux    sync.RWMutex
	metricsClient    metricsclient.Interface
	storage          *Storage
}

func NewMetricsControlller(sdk *k8s.Sdk, metricResolution time.Duration, metriceDuration time.Duration) (*MetricsControlller, error) {

	informerFactory := informers.NewSharedInformerFactory(sdk.ClientSet, 0)
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()
	nodeLister := v1lister.NewNodeLister(nodeInformer.GetIndexer())
	storage, err := NewStorage()
	if err != nil {
		return nil, err
	}
	metricsClient, err := sdk.ToMetricsClient()
	if err != nil {
		return nil, err
	}
	return &MetricsControlller{
		sdk:              sdk,
		informerFactory:  informerFactory,
		nodeInformer:     nodeInformer,
		nodeLister:       nodeLister,
		storage:          storage,
		metricResolution: metricResolution,
		metriceDuration:  metriceDuration,
		metricsClient:    metricsClient,
	}, nil
}

func (s *MetricsControlller) Start(stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start informers
	s.informerFactory.Start(stopCh)
	// go s.nodeInformer.Run(stopCh)

	// Ensure cache is up to date
	ok := cache.WaitForCacheSync(stopCh, s.nodeInformer.HasSynced)
	if !ok {
		return nil
	}
	// Start serving API and scrape loop
	s.runScrape(ctx)

	wait.ContextForChannel(stopCh)
	return nil
}

func (s *MetricsControlller) runScrape(ctx context.Context) {
	ticker := time.NewTicker(s.metricResolution)
	defer ticker.Stop()
	s.tick(ctx, time.Now())

	for {
		select {
		case startTime := <-ticker.C:
			s.tick(ctx, startTime)
		case <-ctx.Done():
			return
		}
	}
}

func (s *MetricsControlller) tick(ctx context.Context, startTime time.Time) {
	s.tickStatusMux.Lock()
	s.tickStatusMux.Unlock()

	ctx, cancelTimeout := context.WithTimeout(ctx, s.metricResolution)
	defer cancelTimeout()

	s.Collect(ctx)

	// s.Store(data)

	// collectTime := time.Since(startTime)
	// tickDuration.Observe(float64(collectTime) / float64(time.Second))
	slog.Info("Scraping cycle complete")
}

func (s *MetricsControlller) Collect(ctx context.Context) {

	nodes, err := s.nodeLister.List(labels.Everything())
	if err != nil {
		slog.Error("List", slog.String("error", err.Error()))
		return
	}
	nodeMetricsList, err := s.metricsClient.MetricsV1beta1().NodeMetricses().List(s.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("List", slog.String("error", err.Error()))
		return
	}

	podMetricsList, err := s.metricsClient.MetricsV1beta1().PodMetricses("default").List(s.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("List", slog.String("error", err.Error()))
		return
	}
	nodeCollect := NewNodeCollect(nodes, nodeMetricsList, podMetricsList)
	nodeCollect.Collect()

	err = s.storage.updateDatabase(nodeMetricsList, podMetricsList)
	if err != nil {
		slog.Error("updateDatabase", slog.String("error", err.Error()))
	}
	slog.Info("Collect", slog.String("success", "true"))
}

func (s *MetricsControlller) GetNodeInnertIp(node *v1.Node) (string, error) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", nil
}
