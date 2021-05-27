# log-exploration-oc-plugin

  To install and run the oc-plugin:
  
  `make install`
  
  To check the installation & fetch all logs:
  
  `oc historical-logs`
  
  Some examples to see how the plugin works:
  
  ```
- Return snapshot historical-logs from pod openshift-apiserver-operator-849d7869ff-r94g8 with a maximum of 10 log extries
oc historical-logs podname=openshift-apiserver-operator-849d7869ff-r94g8 --limit=10

- Return snapshot of historical-logs from pods of stateful set prometheus from namespace openshift-apiserver-operator and logging level info
oc historical-logs statefulset=prometheus --namespace=openshift-apiserver-operator --level=info

- Return snapshot of historical-logs from pods of stateful set nginx in the current namespace with pod name and container name as log prefix
oc historical-logs statefulset=nginx --prefix=true

- Return snapshot of historical-logs from pods of deployment kibana in the namespace openshift-logging with a maximum of 100 log entries
oc historical-logs deployment=kibana --namespace=openshift-logging --limit=100

- Return snapshot of historical-logs from pods of daemon set fluentd in the current namespace
oc historical-logs daemonset=fluentd

- Return snapshot logs of pods in deployment cluster-logging-operator in a time range between current time - 5 minutes and current time
oc historical-logs deployment=cluster-logging-operator --tail=5m

- Return snapshot logs for pods in deployment log-exploration-api in the last 10 seconds
oc historical-logs deployment=log-exploration-api --tail=10s
    
  ```
  
 To run the unit tests:
 
 `make test` 
 
 To check the test-coverage:
 
 `make test-cover`
 
 To run the e2e tests:
 
  optional, if you already have installed elasticdump
  
  `npm install elasticdump`
  
  `make test-e2e`
