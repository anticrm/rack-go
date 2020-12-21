


deploy-docker [image port numInstances] [
  tcp: ip/free-port
  docker/run image port tcp/port
  each-node [kind: load-balancer] [reverse-proxy join tcp/host [":" tcp/port]]
]

each-node [kind: worker] [deploy-docker "screenversation" 3000 os/cpus]

