# Kubernetes Monitoring and Logging Setup Guide

This guide provides detailed instructions on installing and configuring Alertmanager, Loki, Prometheus, and two Grafana instances (development and sandbox) using Helm on your Kubernetes cluster.

## Table of Contents
- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Step 0: Preparations](#step-0-preparations)
- [Step 1: Alertmanager Setup](#step-1-alertmanager-setup)
- [Step 2: Loki Setup](#step-2-loki-setup)
- [Step 3: Prometheus Setup](#step-3-prometheus-setup)
- [Step 4: Grafana Development Environment Setup](#Step-4-grafana-ton-teleport-dev-environment-setup)
- [Final Checks](#final-checks)

## Introduction

This guide will walk you through:
- Installing the required tools (`kubectl` and `helm3`)
- Configuring your Kubernetes context
- Deploying Alertmanager, Loki, Prometheus, and two Grafana instances in separate namespaces

## Prerequisites

Before starting, ensure that you have:
- **kubectl**: Installed and configured to communicate with your Kubernetes cluster.
- **Helm 3**: Installed for managing Helm charts.
- Access to the Kubernetes cluster with the proper permissions.
- Custom values files available in the `.helm_values` directory for each deployment:
  - `alertmagaer.yml` (Note: Verify the file name, it might be intended as `alertmanager.yml`)
  - `loki.yml`
  - `prometheus.yml`
  - `grafana-dev.yml`
  - `grafana-sand.yml`

## Step 0: Preparations

**Install kubectl and helm3**  
   Follow the official installation guides for:
   - [kubectl](https://kubernetes.io/docs/tasks/tools/)
   - [Helm 3](https://helm.sh/docs/intro/install/)

**Configure kubectl**

Save the Configuration File from Rancher web UI (right top corner - Download KubeConfig)
Save your downloaded YAML content to a file (for example, rancher-kubeconfig.yaml).

Set the **KUBECONFIG** Environment Variable
In order for kubectl to use your new configuration file, export the KUBECONFIG variable:

```bash
export KUBECONFIG=./rancher-kubeconfig.yaml
```

Alternatively, if you want to merge this file with your existing kubeconfig, you can use the --kubeconfig flag with kubectl commands.

Select the Context
Although the file sets the current context to develop, you can explicitly switch to it using:

```bash
kubectl config use-context develop
```

Verify the Configuration
To ensure that kubectl is correctly configured, run:

```bash
kubectl get nodes
```

## Step 1: Alertmanager Setup

Add the Prometheus Community Helm repository:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

Create a secret for the Telegram bot token:
This token will be used by Alertmanager to send notifications via Telegram in **ton-teleport-dev**. Replace the placeholder with your actual bot token.

```bash
kubectl create secret generic telegram-bot-token \
  --namespace ton-teleport-dev \
  --from-literal=bot_token="!!!!!!Your raw bot token value here!!!!!!"
```

Deploy Alertmanager using Helm:
This command installs (or upgrades if already installed) the Alertmanager release using your custom values file.

```bash
helm upgrade --install alertmanager prometheus-community/alertmanager \
  --namespace ton-teleport-dev \
  --values ./helm_values/alertmanager.yml
```

## Step 2: Loki Setup

Add the Grafana Helm repository:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Deploy Loki using Helm:
This command installs (or upgrades) Loki in the **ton-teleport-dev** namespace with your custom configuration.

```bash
helm upgrade --install loki grafana/loki \
  --namespace ton-teleport-dev \
  --values ./helm_values/loki.yml
```

## Step 3: Prometheus Setup

Add (or update) the Prometheus Community Helm repository (if not done already):

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

Deploy Prometheus using Helm:
This command installs (or upgrades) Prometheus in your **ton-teleport-dev** namespace with your custom values.

```bash
helm upgrade --install prometheus prometheus-community/prometheus \
  --namespace ton-teleport-dev \
  --values ./helm_values/prometheus.yml
```

## Step 4: Grafana ton-teleport-dev Environment Setup

Add the Grafana Helm repository (if not done already):

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Deploy Grafana (ton-teleport-dev) using Helm:
This command installs (or upgrades) a Grafana instance in the **ton-teleport-dev** namespace.
(the file **grafana-dev.yml** in special for **-dev**)

```bash
helm upgrade --install grafana grafana/grafana \
  --namespace ton-teleport-dev \
  --values ./helm_values/grafana-base.yml \
  --values ./helm_values/grafana-dev.yml
```
## Final Checks

Verify Deployments:
Ensure that all components are running as expected:

```bash
kubectl get pods --all-namespaces
```

Logs and Troubleshooting:
If any pod is not in the Running state, inspect logs with:

```bash
kubectl logs <pod-name> -n <namespace>
```

Helm Releases:
Check the status of your Helm releases:

```bash
helm list --all-namespaces
```