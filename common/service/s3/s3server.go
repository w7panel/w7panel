package s3

import (
	"net/http"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3afero"
)

var S3Http *S3Server

func Init(bucketDir string) {
	S3Http = NewS3Server(bucketDir, "/tmp/metadata", "s3bucket")

}

type S3Server struct {
	directFsBucket string
	proxyServer    *gofakes3.GoFakeS3
}

func NewS3Server(baseFs string, metaFs string, directFsBucket string) *S3Server {
	server := &S3Server{
		directFsBucket: directFsBucket,
	}
	var flags s3afero.FsFlags = 1
	baseFsInit, err := s3afero.FsPath(baseFs, flags)
	if err != nil {
		panic(err)
	}
	metaFsInit, err := s3afero.FsPath(metaFs, flags)
	if err != nil {
		panic(err)
	}

	var backend, err2 = s3afero.SingleBucket(directFsBucket, baseFsInit, metaFsInit)
	if err2 != nil {
		panic(err2)
	}
	server.proxyServer = gofakes3.New(backend,
		gofakes3.WithIntegrityCheck(false),
		// gofakes3.WithTimeSkewLimit(timeSkewLimit),
		// gofakes3.WithTimeSource(timeSource),
		gofakes3.WithLogger(gofakes3.GlobalLog()),
		gofakes3.WithHostBucket(false),
		// gofakes3.WithHostBucketBase(values.hostBucketBases.Values...),
		gofakes3.WithAutoBucket(false),
	)
	return server
}

func (s *S3Server) Server() http.Handler {
	return s.proxyServer.Server()
}

func (s *S3Server) HandlerFunc() http.HandlerFunc {
	return s.proxyServer.Server().ServeHTTP
}

func (s *S3Server) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	s.proxyServer.Server().ServeHTTP(wr, r)
}
