# Operations

## Account

To view all `Account` objects use the command:

```bash
kubectl get accounts.cf.containeroo.ch
```

Output:

```console
NAME   EMAIL                  PHASE
main   admin@example.com      Active
```

## Zone

To view all `Zone` objects use the command:

```bash
kubectl get zones.cf.containeroo.ch
```

Output:

```console
NAME         ZONE NAME    ID                                 PHASE
example-com  example.com  69                                 Active
other-com    other.com    420                                Active
```

## IP

To view all `IP` objects use the command:

```bash
kubectl get ips.cf.containeroo.ch
```

Output:

```console
NAME            ADDRESS         TYPE      PHASE
external-ipv4   178.4.20.69     dynamic   Ready
static-address  142.251.36.35   static    Ready
```

## DNSRecord

To view all `DNSRecord` objects use the command:

```bash
kubectl get dnsrecords.cf.containeroo.ch --all-namespaces
```

Output:

```console
NAMESPACE             NAME                   RECORD NAME            TYPE    CONTENT         PROXIED   TTL   STATUS
blog                  blog-example-com       blog.example.com          A   142.251.36.35    true      1     Created
www                   www-exmaple-com        www.example.com       CNAME   example.com      true      1     Created
cloudflare-operator   vpn-example-com        vpn.example.com       CNAME   example.com      false     120   Created
cloudflare-operator   example-com            example.com               A   178.4.20.69      true      1     Created
```

## Troubleshooting

Usually, cloudflare-operator will store errors in the corresponding object in the `status.message` field and set the object phase (`stattus.phase`) to `Failed`.

### Example

Create a new invalid `DNSRecord`:

```bash
kubectl apply -f - << EOF
apiVersion: cf.containeroo.ch/v1beta1
kind: DNSRecord
metadata:
  name: www-example-com
  namespace: cloudflare-operator
spec:
  name: www.example.com
  type: CNAME
  proxied: true
  ttl: 1
  interval: 5m
EOF
```

!!! info
    The problem is, that `type` is set to `CNAME`, but no `content` is set.

List `DNSRecords`:

```bash
kubectl get dnsrecords --namespace cloudflare-operator www-example-com
```

Output:

```console hl_lines="3"
NAME                   RECORD NAME            TYPE    CONTENT         PROXIED   TTL   STATUS
blog-example-com       blog.example.com          A   142.251.36.35    true      1     Created
www-exmaple-com        www.example.com       CNAME                    true      1     Failed
vpn-example-com        vpn.example.com       CNAME   example.com      false     120   Created
example-com            example.com               A   178.4.20.69      true      1     Created
```

As you can see, the status is `Failed`.

Output the newly created `DNSRecord` to console as YAML:

```bash
kubectl get dnsrecords --namespace cloudflare-operator www-example-com -oyaml
```

Output:

```yaml hl_lines="20 21"
apiVersion: cf.containeroo.ch/v1beta1
kind: DNSRecord
metadata:
  ...
  name: www-example-com
  namespace: cloudflare-operator
  ...
spec:
  interval: 5m
  name: www.example.com
  proxied: true
  ttl: 1
  type: CNAME
status:
  message: DNSRecord content is empty. Either content or ipRef must be set
  phase: Failed
  recordID: ""
```

In the `status.message` you can see the error. The `Phase` is also set to `Failed`.

### Remove Finalizers From All DNSRecords

If you want to remove all finalizers from all`DNSRecords` at once, you can issue the following command:

```bash
kubectl get dnsrecords \
        --all-namespaces \
        --no-headers \
        --output=custom-columns='namespace:.metadata.namespace,name:.metadata.name' | \
  xargs -n 2 \
    kubectl patch dnsrecords.cf.containeroo.ch \
            --patch '{"metadata":{"finalizers":null}}' \
            --type merge \
            --namespace
```

## Metrics

When installing cloudflare-operator with helm, set the following values to enable metrics and Prometheus rules:

```yaml
metrics:
  podMonitor:
    enabled: true
  prometheusRule:
    enabled: true
```

cloudflare-operator then exposes the following metrics:

```text
cloudflare_operator_account_status
cloudflare_operator_dns_record_status
cloudflare_operator_ip_status
cloudflare_operator_zone_status
```

There is also a Grafana dashboard. You can learn more [here](/monitoring).
