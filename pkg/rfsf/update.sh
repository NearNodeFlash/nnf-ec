#!/bin/bash

function usage() {
    echo "Update the rfsf-open API with specified version"
    echo "usage $0 [zip]"
    echo ""
    echo "  zip file residing in https://www.dmtf.org/sites/default/files/standards/documents/"
    echo "  or local copy in $(pwd)"
    exit 1
}

function repackage() {
    sed -i '' -e "s/package openapi/package ${1}/" $2
}

if [ "$1" == "" ]; then
    usage
fi

DIR="${1%.*}"
YAML="yaml/openapi.yaml"
OUT="out"

echo "Using: $YAML"
echo "Output Directory: $OUT"

if [ ! -f $YAML ]; then
    echo "yaml file not found"

    if [ ! -f $1 ]; then
        echo "Fetching $1"
        wget https://www.snia.org/sites/default/files/technical_work/PublicReview/swordfish/$1
    fi

    echo "Unzip $YAML"
    unzip -u $1 $YAML

    echo "Patching the $YAML file"
    mv $YAML $YAML.orig
    ./tools/patch_yaml.py $YAML.orig > $YAML
fi

echo "Running OpenAPI generator on $YAML"
OPTS='--type-mappings=integer=int64'
openapi-generator generate -g go-server -i $YAML -o $OUT $OPTS

echo "Moving generated files from $OUT to package tree"
cp -r $OUT/.openapi-generator ./
cp -r $OUT/api ./
rm -rf ./pkg/models/*.go && cp $OUT/go/model_* ./pkg/models/
cp $OUT/go/logger.go ./pkg/logger/ && repackage logger ./pkg/logger/logger.go

cp $OUT/go/api.go ./pkg/routermux && repackage routermux ./pkg/routermux/api.go
cp $OUT/go/api_default.go ./pkg/routermux && repackage routermux ./pkg/routermux/api_default.go
cp $OUT/go/api_default_service.go ./pkg/routermux && repackage routermux ./pkg/routermux/api_default_service.go
cp $OUT/go/routers.go ./pkg/routermux && repackage routermux ./pkg/routermux/routers.go

echo "Patching OEM fields"
sed -i '' -e 's/Oem map\[string\]/Oem /g' ./pkg/models/model_*.go

echo "Patching  pointer fields"
sed -i '' -e 's/\*string/string/g' ./pkg/models/model_*.go
sed -i '' -e 's/\*int/int/g'       ./pkg/models/model_*.go
sed -i '' -e 's/*bool/bool/g'      ./pkg/models/model_*.go


echo "Patching routermux openapi imports"
./tools/patch_open_api_models.py ./pkg/routermux --file api.go
./tools/patch_open_api_models.py ./pkg/routermux --file api_default.go
./tools/patch_open_api_models.py ./pkg/routermux --file api_default_service.go

echo "Patching generated Go constants"
./tools/patch_generated_constants.py ./pkg/models

echo "Generating Storage Platform API endpoints"
./tools/storage_platform_generator.py sp_api.go api_default.go sp_api_default.go ./pkg/routermux
gofmt -w pkg/routermux/sp_api_default.go

