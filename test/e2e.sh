#!/bin/bash

set -e

# Bin path
BIN_PATH=$1
if [ "${BIN_PATH}" = "" ]
then 
  echo "must specify the bin path of registy"
  exit -1
fi

# Project path
BIN_ROOT="$(cd $(dirname "${BIN_PATH}");pwd)"
E2E_ROOT="$(cd $(dirname "${BASH_SOURCE}");pwd)"
TMP_PATH=${BIN_ROOT}/temp
DATA_PATH=${TMP_PATH}/data

# Clean up 
function local-cleanup {
  [ -n "${REGISTRY_PID-}" ] && ps -p ${REGISTRY_PID} > /dev/null && kill ${REGISTRY_PID}
  rm -rf ${TMP_PATH}
}
trap local-cleanup INT EXIT


# Run registry
mkdir -p ${DATA_PATH}
cat > ${TMP_PATH}/config.yaml <<EOF
listen: ":9999"
manager: 
  name: "simple"
  parameters: 
    storagedriver: filesystem
    rootdirectory: "${DATA_PATH}"
EOF
${BIN_PATH} serve -c ${TMP_PATH}/config.yaml &
REGISTRY_PID=$!


# Set up environment
export ENV_ENGPOINT="http://127.0.0.1:9999"

# Run tests
testcase=(
space
chart
)

cd ${E2E_ROOT}
for case in ${testcase[@]}
do
  go test -race -v ./${case}
  if [ $? -ne 0 ]
  then
    cd -
    exit -1
  fi
done
cd -
