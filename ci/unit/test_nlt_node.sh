#!/bin/bash

# This is a script to be run by the ci/unit/test_nlt.sh to run a test
# on a CI node.

set -uex

sudo bash -c 'echo 1 > /proc/sys/kernel/sysrq'
sudo mkdir -p /mnt/daos
# using mmap()'ed ULT stacks requires to bump system default
if [ "$(sudo sysctl -n vm.max_map_count)" -lt "1000000" ] ; then
    sudo sysctl vm.max_map_count=1000000
fi

sudo mkdir /opt/daos
sudo chmod 777 /opt/daos

cd build

ls
cp -a opt-daos /opt/daos/
find /opt/daos

# Setup daos admin etc.
sudo bash -c ". ./utils/sl/setup_local.sh; ./utils/setup_daos_admin.sh"

# NLT will mount /mnt/daos itself.
./utils/node_local_test.py --max-log-size 650MiB --dfuse-dir /localhome/jenkins/ all
