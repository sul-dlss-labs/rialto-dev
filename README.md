# rialto-dev
A docker-based development environment for RIALTO

## Requirements
* git
* ruby
* node
* docker
* docker-compose
* golang
* python3
* pip
* AWS cli

## Setup
1. Clone the repos
    ```shell
    mkdir rialto
    cd rialto
    git clone https://github.com/sul-dlss-labs/rialto-dev.git
    git clone https://github.com/sul-dlss/rialto-etl.git
    git clone https://github.com/sul-dlss/rialto-webapp.git
    git clone https://github.com/sul-dlss/sparql-loader.git
    mkdir -p $GOPATH/src/github.com/sul-dlss
    cd $GOPATH/src/github.com/sul-dlss
    git clone https://github.com/sul-dlss/rialto-derivatives.git
    git clone https://github.com/sul-dlss/rialto-entity-resolver.git
    ```
1. Setup rialto-dev
    ```shell
    cd rialto-dev
    mkdir bg-data
    mkdir solr-data
    mkdir postgres-data
    cp example.env .env
    ```

    If there are port conflicts with existing applications, you can change the ports
used by the dev environment in `.env`.
    
1. Setup rialto-etl  
    
    Add `config/local.settings.yml`. In addition to actual API keys, this should include:
  
    ```yaml
    sparql_writer:
      update_url: http://localhost:8082/sparql
      batch_size: 1000

    entity_resolver:
      url: http://localhost:3000/
      api_key: abc123
    ```
    
1. Setup rialto-webapp

    ```shell
    cd rialto-webapp
    bundle install
    npm install
    ```
    
1. Setup sparql-loader  

    First, [install dependencies](https://github.com/sul-dlss/sparql-loader#install-dependencies) such as Python, pip, and virtualenv. Prefer Python >= 3. Then:
  
    ```shell
    cd sparql-loader
    virtualenv -p $(which python3) env
    source env/bin/activate
    pip install -r requirements.txt
    ```
    
1. Deploy rialto-derivatives
    
    ```shell
    cd $GOPATH/src/github.com/sul-dlss/rialto-derivatives
    dep ensure
    make local-deploy
    ```

## Starting dev environment

```shell
cd rialto-dev
docker-compose up -d
```

Note: The first time you run this, it will take some time to download and build images.

Rialto and components will now be available at:
* Blazegraph: http://localhost:9999/blazegraph/
* Localstack UI: http://localhost:8081/
* Postgres: port 5432
* Entity resolver: port 3000
* Solr: http://localhost:8983/solr/
* Sparql-loader: http://localhost:8082/sparql
* Webapp: http://localhost:8080/

## Deploying derivative lambdas

This should be performed after the dev environment has been started or whenever
the localstack container has been restarted.

```shell
cd $GOPATH/src/github.com/sul-dlss/rialto-derivatives
make local-deploy
```

This will compile and zip the lambdas, delete them in Localstack if they already exist,
create them in Localstack, create a topic, and subscribe the lambdas to the topic.

Separate make commands exist for each of these steps.

## Loading data

Assuming that you have already extracted data and converted to sparql.

Because this is a local development environment, it is recommended to only load
a sample of data.

```shell
exe/load call Sparql -i organizations.sparql
exe/load call Sparql -i researchers_sample.sparql
```

Don't forget to cleanup containers with `docker container prune -f`.

## Derivative lambda development

Changes will need to be deployed to Localstack using `make local-deploy`.

## Webapp development

Since your local code is linked into the container, changes will automatically cause
the running webapp to be reloaded.

The first time that you invoke the webapp, there may be a wait as javascript is packed.

If you change any gem files, rebuild the image with:

```shell
docker-compose stop webapp
docker-compose build webapp
docker-compose up -d
```

## Sparql loader development

Local code is not linked into the container. To reload the server, the image must be rebuild.

```shell
docker-compose stop sparql-loader
docker-compose build sparql-loader
docker-compose up -d
```

## Notes

### Connecting to the db

```shell
docker exec -it rialto-dev_db_1 psql -h db -U postgres -d rialto_development
```

### Cleaning up docker containers

Localstack creates a new container each time a lambda is executed. To clean up these containers:

```shell
docker container prune -f
```

### Clearing state

State for Blazegraph, Postgres, and Solr are persisted. (State for Localstack is not.)

To clear state:

```shell
rm -fr bg-data/*
rm -fr postgres-data/*
rm -fr solr-data/*
```

## Helpful docker commands

* Stop all containers: `docker-compose stop`
* Stop single container: `docker-compose stop webapp`
* Remove all containers: `docker-compose rm -vf`
* Remove all stopped containers (not just specified in `docker-compose.yml`): `docker container prune`
* View and follow the log of a container: `docker-compose logs -f webapp`
* Rebuild a container: `docker-compose build webapp`
* View all containers, including stopped: `docker ps -a`
* View the logs of a container by name: `docker logs jolly_kepler`
* Shell into a container: `docker exec -it rialto-dev_webapp_1 /bin/bash`

## Kludges

* Because of limitations in Localstack, the sparql loader lambda is run as a server in
  a docker container.
* Because Localstack [does not attach the lambda docker containers to the correct network](https://github.com/localstack/localstack/issues/381),
  the `docker-events-listener` container attachs them open creation. This introduces
  a latency in the lambdas being able to reach the triplestore, which is handled
  by retrying the connection.
* When executing `db:setup`, [rails ignores the environment](https://github.com/rails/rails/issues/27299)
  and tries to create both the development and test environment. This error is
  ignored in `invoke.sh`.
