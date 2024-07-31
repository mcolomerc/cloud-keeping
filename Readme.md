# Cloud Cleaning Tool

<img src="./docs/logo.jpeg " width="100" height="100">

`cleanup` is a tool to clean up Cloud resources.

## Confluent Cloud

### Topics

Deletes all inactive topics in a cluster in Confluent Cloud. It uses Confluent Cloud Metrics API to get the list of topics using `io.confluent.kafka.server/received_records` metric. If the topic has not received any messages in the last 7 days, it is considered inactive and will be deleted.

#### Usage

Confluent Cloud API Key and Secret are required to access the metrics API. Cluster API Key and Secret are required to delete the topics.

- [Confluent Cloud Metrics API](https://docs.confluent.io/cloud/current/metrics-api.html)
- [Confluent Cloud API Keys](https://docs.confluent.io/cloud/current/security/api-keys.html)
- [Confluent Cloud Cluster API Keys](https://docs.confluent.io/cloud/current/security/api-keys.html#cluster-api-keys)

- Using flags:

```shell
Usage:
  cleanup confluent topics [flags] 

Flags:
      --cloud_api_key string        Cloud API KEY with Metrics API access or set CLOUD_API_KEY environment variable
      --cloud_api_secret string     Cloud API SECRET or set CLOUD_API_SECRET environment variable
      --cluster string              A Confluent Cloud cluster Id (lkc-xxxxx) or set CLUSTER environment variable
      --cluster_api_key string      Cluster API KEY or set CLUSTER_API_KEY environment variable
      --cluster_api_secret string   Cluster API SECRET or set CLOUD_API_KEY environment variable
      --environment string          Confluent Cloud environment Id (env-xxxxx) or set ENVIRONMENT environment variable
  -h, --help                        help for topics
```

- Using environment variables:

```shell
export ENVIRONMENT=env-<id>
export CLUSTER=lkc-<id>
export CLUSTER_API_KEY=<CLUSTER_API_KEY>
export CLUSTER_API_SECRET=<CLUSTER_API_SECRET>
export CLOUD_API_KEY=<CLOUD_API_KEY>
export CLOUD_API_SECRET=<CLOUD_API_SECRET>

cleanup confluent topics
```

Internal topics are not considered for deletion.
Ouput will be a list of topics and a prompt to delete all inactive topics:

```shell
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────────────────╮
│ TOPIC                                                                                                             │ ACTIVE (LAST 7 DAYS) │
├───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
│ _confluent-ksql-pksqlc-kn8m66query_CSAS_STOCKS_ENRICHED_21-Join-repartition                                       │ YES                  │
│ _confluent-ksql-pksqlc-kn8m66query_CSAS_STOCKS_ENRICHED_21-KafkaTopic_Right-Reduce-changelog                      │ YES                  │
│ pksqlc-kn8m66TOTAL_STOCK_PURCHASED                                                                                │ YES                  │
│ stocks_topic                                                                                                      │ YES                  │
│ topic_inactive_3                                                                                                  │ EMPTY                │
│ users_topic                                                                                                       │ YES                  │
├───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────────────┤
│ TOTAL                                                                                                             │ 28                   │
╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────────────────╯
? Delete all inactive topics?? [y/N] █
```

Confirm deletion:

```shell
Delete all inactive topics?: y
╭──────────────────┬─────────╮
│ TOPIC            │         │
├──────────────────┼─────────┤
│ topic_inactive_3 │ DELETED │
├──────────────────┼─────────┤
│ TOTAL            │ 1       │
╰──────────────────┴─────────╯
```

