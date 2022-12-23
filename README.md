## Problem Statement

Deploy a system that can get users download links for the albums they purchase.

-   A user comes in with a download key
-   The key form/validity must be checked
-   The key cannot have been used before
-   We need to provide a link to the user

## Solution Description

This solution uses:

-   helm and k8s on Rancher Desktop
-   elasticsearch for the database
-   golang-based web server inside a container controlled by a k8s Deployment

At least at first, I won't separate the functions across processes, as I want my end user to leave with either an explanation about why his ticket is rejected, or the link to download this swag.

## One-time Setup

Before doing any of this, have `kubectl` pointed at a suitable k8s cluster.

    kubectl config current-context # should be healthy e.g. rancher-desktop

Set up the Elasticsearch helm repo:

    helm repo add elastic https://helm.elastic.co
    helm repo update

Ensure you have a storage class defined called `local-path` that the k8s statefulset needs or update TODO

    kubectl get storageclass

## Install Procedure

Within the project directory:

    kubectl create namespace tickey-lickey
    helm install elasticsearch elastic/elasticsearch -n tickey-lickey -f elasticsearch-values.yaml
    alias kct='kubectl -n tickey-lickey' # make things easier

Test cluster health using Helm test.

    helm --namespace=tickey-lickey test elasticsearch

Port-forward 9200 and hit the elastic endpoint:

    kct port-forward service/elasticsearch-master 9200:9200 &
    ESPASS=$(kct get secrets elasticsearch-master-credentials -ojsonpath='{.data.password}' | base64 -d)
    http -a elastic:$ESPASS --verify=no https://localhost:9200/ # brew install httpie

You should see 3 elasticsearch pods, 2 services and 1 deployment:

    kct get all

## Build the Ticket Redemption Go Application

In order for Rancher k8s to use the images you build, you must build them with nerdctl (assuming you're using containerd?) and put them in the k8s.io namespace:

    nerdctl build --namespace k8s.io -t tickey-lickey-redeemer:latest ./apps/redeemer

## Deploy the Redeemer Server Chart and Health Check Service

    helm install -n tickey-lickey redeemer ./redemption-server
    kct port-forward service/redeemer 8080:8080 &
    http localhost:8080 # Hello?

## Load Ticket Data

    http --verify=no POST https://localhost:9200/tickets/_bulk -a elastic:$ESPASS -j < redemption-server/ticket-data.json

Query the tickets that are there:

    cat <<EOF | http --verify=no 'https://localhost:9200/_search?pretty' -a elastic:$ESPASS -j
    {
        "query": { "match_all": {} }
    }
    EOF

Now you can try to redeem tickets:

    http://localhost:8080/redeem/youtoowonaprize
