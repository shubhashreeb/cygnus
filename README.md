# cygnus

exposing service through cygnus

apiVersion: v1
kind: Service
metadata:
  name: svc1
  annotations:
    cygnus.io/config: |
      ---
      apiVersion: cygnus/v1
      kind: Mapping
      name: svc1-service_mapping
      host: svc1.your-domain
      prefix: /
      service: svc1:80
spec:
  selector:
    app: nginx
    name: svc1
  ports:
  - name: http
    protocol: TCP
    port: 80



https://www.digitalocean.com/community/tutorials/how-to-create-an-api-gateway-using-ambassador-on-digitalocean-kubernetes


K8s Networking

https://www.getambassador.io/docs/emissary/latest/topics/concepts/kubernetes-network-architecture/

