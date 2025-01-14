# cloudflare-operator Documentation

The goal of cloudflare-operator is to manage Cloudflare DNS records using Kubernetes objects.

cloudflare-operator is built from the ground up to use Kubernetes' API extension system.

## Who is cloudflare-operator for?

cloudflare-operator helps to:

- Manage Cloudflare DNS records using Kubernetes objects
- Keep Cloudflare DNS records up to date
- Update your external IP address on Cloudflare DNS records

## What can I do with cloudflare-operator?

cloudflare-operator is based on a set of Kubernetes API extensions ("custom resources"), which control Cloudflare DNS records.

## Where do I start?

Following [this](getting_started.md) guide will just take a couple of minutes to complete. After installing the cloudflare-operator helm chart and adding some annotation to your ingresses, cloudflare-operator will take care of your Cloudflare DNS records.

## More detail on what’s in cloudflare-operator

Features:

- Add, update and delete Cloudflare DNS records
- Update Cloudflare DNS records if your external IP address changes
