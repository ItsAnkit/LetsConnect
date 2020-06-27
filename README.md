# LetsConnect

### Install required go packages:
  > go build

### Setup Cassandra
  > brew install ccm maven
  > pip install cqlsh
  
### Create cassandra cluster
  > ccm create -v 3.9 letsconnect
  
### Setup nodes in cluster
  > ccm populate -n 1
  
### Start cassandra server
  > ccm start
  
### Create Database
  > sh cassandra.sh
  
Note: You need jdk installed in your system to run cassandra.
  
