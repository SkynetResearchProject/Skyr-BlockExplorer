#!/bin/bash
#SKYR testnet
#bin/blockbook -sync -resyncindexperiod=60017 -resyncmempoolperiod=60017 -blockchaincfg=build/tnblockchaincfg.json -internal=:10131 -public=:443  -logtostderr
#SKYR mainnet
#bin/blockbook  -sync -resyncindexperiod=60017 -resyncmempoolperiod=60017 -blockchaincfg=build/blockchaincfg.json -internal=:10130 -public=:443  -logtostderr
#docker
bin/blockbook  -sync -resyncindexperiod=60017 -resyncmempoolperiod=60017 -blockchaincfg=build/blockchaincfg.json -internal=:10030 -public=:10130  -logtostderr
