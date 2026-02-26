package metrics

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func CreateDatabase(db *sql.DB) error {
	sqlStmt := `
	create table if not exists nodes (uid text, name text, cpu text, memory text, storage text, hostcoreutilization text,hostgpumemoryusage text, time datetime);
	create table if not exists pods (uid text, name text, namespace text, container text, cpu text, memory text, storage text, time datetime);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

type Storage struct {
	db *sql.DB
}

func NewStorage() (*Storage, error) {
	dbFile := facade.Config.GetString("metrics.db_file")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		// log.Fatalf("Unable to open Sqlite database: %s", err)
		return nil, fmt.Errorf("Unable to open Sqlite database: %s", err)
	}
	// defer db.Close()
	CreateDatabase(db)
	return &Storage{db}, nil
}

func (self *Storage) updateDatabase(nodeMetrics *v1beta1.NodeMetricsList, podMetrics *v1beta1.PodMetricsList) error {

	db := self.db
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into nodes(uid, name, cpu, memory, storage, time, hostcoreutilization, hostgpumemoryusage) values(?, ?, ?, ?, ?, datetime('now', 'localtime'), ?, ?)")
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range nodeMetrics.Items {
		val2, ok := v.Usage[HostCoreUtilizationName]
		hostCoreVal := "0"
		if ok {
			hostCoreVal = val2.String()
		}
		val3, ok := v.Usage[HostGPUMemoryUsageName]
		HostUsage := "0"
		if ok {
			HostUsage = val3.String()
		}
		_, err = stmt.Exec(v.UID, v.Name, v.Usage.Cpu().MilliValue(), v.Usage.Memory().MilliValue()/1000, v.Usage.StorageEphemeral().MilliValue()/1000, hostCoreVal, HostUsage)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	stmt2, err := tx.Prepare("insert into pods(uid, name, namespace, container, cpu, memory, storage, time) values(?, ?, ?, ?, ?, ?, ?, datetime('now', 'localtime'))")
	defer stmt2.Close()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range podMetrics.Items {
		for _, u := range v.Containers {
			_, err = stmt2.Exec(v.UID, v.Name, v.Namespace, u.Name, u.Usage.Cpu().MilliValue(), u.Usage.Memory().MilliValue()/1000, u.Usage.StorageEphemeral().MilliValue()/1000)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	err = tx.Commit()

	if err != nil {
		rberr := tx.Rollback()
		if rberr != nil {
			return rberr
		}
		return err
	}

	return nil
}

/*
CullDatabase deletes rows from nodes and pods based on a time window.
*/
func (self *Storage) CullDatabase(window time.Duration) error {
	db := self.db
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	windowStr := fmt.Sprintf("-%.0f seconds", window.Seconds())

	nodestmt, err := tx.Prepare("delete from nodes where time <= datetime('now', ?);")
	if err != nil {
		tx.Rollback()
		return err
	}

	defer nodestmt.Close()
	res, err := nodestmt.Exec(windowStr)
	if err != nil {
		tx.Rollback()
		return err
	}

	affected, _ := res.RowsAffected()

	podstmt, err := tx.Prepare("delete from pods where time <= datetime('now', ?);")
	defer podstmt.Close()
	if err != nil {
		tx.Rollback()
		return err
	}

	res, err = podstmt.Exec(windowStr)
	if err != nil {
		tx.Rollback()
		return err
	}

	affected, _ = res.RowsAffected()
	klog.V(klog.Level(0)).Infof("Cleaning up pods: %d rows removed", affected)
	err = tx.Commit()

	if err != nil {
		rberr := tx.Rollback()
		if rberr != nil {
			return rberr
		}
		return err
	}

	return nil
}
