from __future__ import print_function
import sys
import argparse
import json

DOMAIN = "trustingsocial.com"
SRE = "sre"

def user_ldap_config(one_user_object, user_id, group_id):

    email = one_user_object.get("email")
    user_name = one_user_object.get("name")

    if user_name and "@" in user_name:
        user_name = email.split("@")[0]

    user = {}
    user["name"] = user_name
    user["givenname"] = user_name
    user["otpsecret"] = one_user_object.get("otp_secret","")
    user["mail"] = email
    user["homeDir"] = "/home/%s" % user_name
    user["loginShell"] = "/bin/bash"
    user["primarygroup"] = int(group_id)
    user["unixid"] = int(user_id)
    return user

def get_user_group_id(one_user_object, desired_groups):
    group_id = None
    is_sre = False
    user_id = None

    for i in desired_groups.keys():
        for g in one_user_object.get("groups"):
            if i == SRE and i == g:
                group_id = desired_groups[i]["group_id"]
                user_id = desired_groups[i]["user_id"]
                is_sre = True
                continue

            if not is_sre and i == g:
                group_id = desired_groups[i]["group_id"]
                user_id = desired_groups[i]["user_id"]

    return user_id, group_id

def add_users(data, desired_groups):
    users = []

    for u in data:
        email = u.get("email")
        if u.get("disabled") \
            or not email \
            or not email.endswith(DOMAIN):
            continue

        uid, gid = get_user_group_id(u, desired_groups)
        if not gid:
            continue

        ret = user_ldap_config(u, uid, gid)
        if not ret:
            continue
        users.append(ret)
    return users

def add_base_dn():
    return "dc=tsocial,dc=com"

def add_groups(desired_groups):
    if SRE not in desired_groups.keys():
        raise Exception(
            "No 'sre' group found in permission config! Cannot proceed"
        )

    groups = []
    for g,i in desired_groups.iteritems():
        groups.append({"name": g, "unixid": int(i.get("group_id"))})
    return groups

def parse(data, acl_config):
    _final = {}

    desired_groups = acl_config["internal"]
    data.append(acl_config["external"]["users"]["searcher"])

    _final["groups"] = add_groups(desired_groups)
    _final["baseDN"] = add_base_dn()
    _final["users"] = add_users(data, desired_groups)
    return _final

def dump_to_stdout(output):
    print(json.dumps(json.loads(json.dumps(output)),
                     indent=4,
                     sort_keys=True))

def parser():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-d", "--data",
        help="Input Pritunl JSON data file",
        required=True
    )

    parser.add_argument(
        "-c", "--acl_config",
        help="Input file with User ID, Group ID or external users",
        required=True
    )

    return parser

def load(args):
    with open(args.data, "r") as d:
         data = json.load(d)

    with open(args.acl_config, "r") as c:
         desired_groups = json.load(c)

    return data, desired_groups

if __name__ == "__main__":
    args = parser().parse_args()
    data, acl_config = load(args)
    dump_to_stdout(
        parse(data, acl_config)
    )
