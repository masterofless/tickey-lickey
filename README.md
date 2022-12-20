## Problem Statement

Deploy a system that can get users download links for the albums they purchase.

A user comes in with a download key.

The key form/validity must be checked.

The key cannot have been used before.

We need to provide a link to the user.

## Solution Description

I plan on using

-   helm and k8s on Rancher Desktop
-   go
-   elasticsearch for the database

At least at first, I won't separate the functions across processes, as I want my end user to leave with either an explanation about why his ticket is rejected, or the link to download this swag.

## One-time Setup

Before doing any of this, have `kubectl` pointed at a suitable k8s cluster.

Set up the Elasticsearch help repo:

    helm repo add elastic https://helm.elastic.co
    helm repo update

Ensure you have a storage class defined called `local-path` that the k8s statefulset needs or update TODO

## Install Procedure

Within the project directory:

    kubectl create namespace tickey-lickey
    helm install elasticsearch elastic/elasticsearch -n tickey-lickey -f values.yaml
    alias kct='kubectl -n tickey-lickey' # make things easier

Test cluster health using Helm test.

    helm --namespace=tickey-lickey test elasticsearch

and port-forward 9200 and hit the elastic endpoint:

    kct port-forward service/elasticsearch-master 9200:9200 &
    ESPASS=$(kct get secrets elasticsearch-master-credentials -ojsonpath='{.data.password}' | base64 -d)
    http -a elastic:$ESPASS --verify=no https://localhost:9200/ # brew install httpie

You should see 3 elasticsearch pods, 2 services and 1 deployment:

    kct get all

## Build the Ticket Redemption Go Application

    nerdctl build -t tickey-lickey-redeemer:latest --namespace tickey-lickey go_server

## Deploy the Redeemer Server Chart and Health Check Service

    helm install -n tickey-lickey redeemer ./redemption-server
    kct port-forward service/redeemer 8080:8080
    http localhost:8080 # Hello?

## Load Ticket Data

    http --verify=no POST https://localhost:9200/tickets/_bulk -a elastic:$ESPASS -j < redemption-server/ticket-data.json

Query the tickets that are there:

    cat <<EOF | http --verify=no 'https://localhost:9200/_search?pretty' -a elastic:$ESPASS -j
    {
        "query": {
            "match_all": {}
        }
    }
    EOF
