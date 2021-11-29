#!/bin/bash

if [ $# != 3 ]; then
    echo "getNW.sh> ([1] profile) ([2] Tag Name) ([3] Tag Value)"
    exit 1
fi

InstansDetail=`aws ec2 describe-instances --filters "Name=tag:$2,Values=$3" --profile $1`
SubnetID=`echo $InstansDetail | jq .Reservations[].Instances[].SubnetId | tr -d "\""`
SubnetName=`aws ec2 describe-subnets --profile $1 --filters "Name=subnet-id,Values=$SubnetID" | jq .Subnets[].Tags[].Value | tr -d "\""`

SG=`echo $InstansDetail | jq -r ".Reservations[].Instances[].SecurityGroups[] | [.GroupName, .GroupId] | @csv" | tr -d "\""`

echo $SubnetName,$SubnetID
echo $SG
