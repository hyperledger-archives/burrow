module github.com/hyperledger/burrow

go 1.12

replace github.com/go-interpreter/wagon v0.0.0 => github.com/perlin-network/wagon v0.3.1-0.20180825141017-f8cb99b55a39

// replace github.com/tendermint/tendermint => github.com/tendermint/tendermint master
replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.32.2-0.20190919084157-4dfbaeb0c412

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/OneOfOne/xxhash v1.2.5
	github.com/alecthomas/jsonschema v0.0.0-20190122210438-a6952de1bbe6
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf // indirect
	github.com/btcsuite/btcd v0.0.0-20190418232430-6867ff32788a
	github.com/cep21/xdgbasedir v0.0.0-20170329171747-21470bfc93b9
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elgs/gojq v0.0.0-20160421194050-81fa9a608a13
	github.com/elgs/gosplitargs v0.0.0-20161028071935-a491c5eeb3c8 // indirect
	github.com/fatih/color v1.7.0
	github.com/go-kit/kit v0.9.0
	github.com/go-ozzo/ozzo-validation v3.5.0+incompatible
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/golang-lru v0.5.1
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/imdario/mergo v0.3.7
	github.com/jawher/mow.cli v1.1.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.1.1
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/monax/relic v2.0.0+incompatible
	github.com/perlin-network/life v0.0.0-20190521143330-57f3819c2df0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.4.0
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/viper v1.4.0
	github.com/streadway/simpleuuid v0.0.0-20130420165545-6617b501e485
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/iavl v0.12.4
	github.com/tendermint/tendermint v0.32.3
	github.com/tendermint/tm-db v0.1.1
	github.com/tmthrgd/go-hex v0.0.0-20190303111820-0bdcb15db631
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.1.0
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys v0.0.0-20190825160603-fb81701db80f // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190826060629-95c3470cfb70 // indirect
	google.golang.org/grpc v1.23.1
	gopkg.in/yaml.v2 v2.2.2
)
