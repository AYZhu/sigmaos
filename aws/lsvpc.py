#!/usr/bin/env python
import argparse

import boto3

parser=argparse.ArgumentParser(description='ls vpc on AWS.')
parser.add_argument('vpc-id', help='VPC')
args = vars(parser.parse_args())

def ls_nets(vpc):
    subnets = list(vpc.subnets.all())
    if len(subnets) > 0:
        for sn in subnets:
            print("Subnet:", sn.id, "-", sn.cidr_block)
    else:
        print("There is no subnet in this VPC")

def ls_sg(vpc):
    for sg in vpc.security_groups.all():
        if sg.group_name != 'default':
            print("Security group:", sg.id, sg.group_name)

def name(tags):
    name = ""
    for d in tags:
        if d['Key'] == 'Name':
            name = d['Value']
    return name

def ls_instances(vpc):
    client = boto3.client('ec2')
    response = client.describe_instances(Filters=[
        {"Name": "vpc-id", "Values": [vpc.id]}
    ])
    vms = []
    for r in response['Reservations']:
        for i in r['Instances']:
            if 'Tags' in i:
                vms.append((i['InstanceId'], i['PublicDnsName'], i['Tags']))
            else:
                vms.append((i['InstanceId'], i['PublicDnsName'], []))
    if vms == []:
        print("There is no instance in this VPC")
    else:
        for vm in vms:
            print("VMInstance", name(vm[2]), ":", vm[0], vm[1])

def ls_db(vpc):
    client = boto3.client('rds')
    response = client.describe_db_instances()
    rds_instances = list(filter(lambda x: x["DBSubnetGroup"]["VpcId"] == vpc.id, response["DBInstances"]))
    if len(rds_instances) > 0:
        for rds in rds_instances:
            print("RDSInstance:", rds['DBInstanceIdentifier'], rds['Endpoint']['Address'])
    else:
        print("There is no RDS instance in this VPC")

def main():
   vpc_id = args['vpc-id']

   boto3.setup_default_session(profile_name='me-mit')
   ec2 = boto3.resource('ec2')

   try:
       vpc = ec2.Vpc(vpc_id)
       print("VPC Name:", name(vpc.tags))
       ls_nets(vpc)
       ls_sg(vpc)
       ls_instances(vpc)
       ls_db(vpc)

   except Exception as e:
       print("error", e)


if __name__ == "__main__":
    main()
    
