


deploy-image: fn [deploymentUrl image port numInstances] [
  tcp: ip/free-port
  docker/run image port tcp/port
  cluster/each-node [kind: load-balancer] [reverse-proxy deploymentUrl join tcp/host [":" tcp/port]]
]

deploy: fn [deploymentUrl buildId] [
  cluster/each-node [kind: worker] [deploy-image deploymentUrl join "anticrm/scrn:" buildId 3000 os/cpus]
]

cluster/deploy [deploy join os/args/1 ".screenversaion.com" env/BUILD_ID]
