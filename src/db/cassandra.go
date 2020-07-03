package db

import (
	"github.com/gocql/gocql"
	"log"
)

var Session *gocql.Session

func StartCassandraSession() *gocql.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "letsconnect"
	//cluster.Consistency = gocql.LocalQuorum
	Session, err := cluster.CreateSession()
	if err != nil {
		log.Println("Cassandra Session Error: ", err)
	}
	return Session
}