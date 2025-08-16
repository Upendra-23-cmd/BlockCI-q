-------------------------MY FILE STRUCTURE FOR THE PROJECT----------------------------
blockci-q/
│
├── cmd/                     # entrypoints for binaries
│   ├── agent/               # worker/agent node (executes jobs)
│   ├── server/              # API server (webhook, jobs, ledger)
│   └── tester/main.go       # test runner for dev
│
├── configs/                 # YAML configs for server & agent
│   ├── agent.yaml
│   └── server.yaml
│
├── deployments/             # deployment artifacts
│   ├── docker-compose.yml   # local multi-service setup
│   ├── dockerfile.agent     # build agent container
│   ├── dockerfile.server    # build server container
│   └── k8s/                 # Kubernetes manifests
│
├── internal/                # main application logic
│   ├── agent/               # agent runtime (executes steps/jobs)
│   ├── api/                 # REST/GraphQL API server
│   ├── blockchain/          # immutable ledger / audit logs
│   ├── core/                # core CI/CD engine
│   │   ├── job.go           # job definition
│   │   ├── parser.go        # YAML parser
│   │   ├── pipeline.go      # pipeline + stages
│   │   └── scheduler.go     # scheduler logic
│   ├── security/            # authn/authz, SPIFFE/OPA later
│   └── storage/             # artifacts, logs, results
│
├── pkg/                     # reusable libs (can be imported by internal/*)
│   ├── config/              # config loader
│   ├── logger/              # logging abstraction
│   └── utils/               # helpers
│
├── scripts/                 # devops scripts
│   ├── build.sh             # build binaries/images
│   ├── migrate.sh           # DB migrations (ledger/artifacts DB)
│   └── run_local.sh         # run local stack
│
├── tests/                   # integration/unit tests
│
├── web/                     # (future) dashboard UI
│
├── pipeline.yaml            # sample pipeline definition
├── go.mod
└── go.sum


-------------------------------------------------------------------------------------

