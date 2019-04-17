import unittest
import argparse
from generate import parse, load
from unittest import TestCase

class TestConfigGen(TestCase):
    def setUp(self):
        attrs = self.setattrs()
        self.data, self.acl_config = load(attrs)
        self.final_config = parse(self.data, self.acl_config)

    def setattrs(self):
        '''
        Setup the attributes
        '''
        parser = argparse.ArgumentParser(description="Argument Parser")
        attrs = parser.parse_args()
        setattr(attrs, "data", "test_data/pritunl_users_sample.json")
        setattr(attrs, "acl_config", "test_data/permission_config.json")
        return attrs

    def test_group(self):
        '''
        Test group length
        '''
        self.assertEquals(len(self.final_config["groups"]), 2)
        for g in self.final_config["groups"]:
            self.assertTrue(
                g.get("name") in ["tks", "sre"]
            )

    def test_len_elligible_user(self):
        '''
        Test length of elligible users
        '''
        self.assertEquals(len(self.final_config["users"]), 3)

    def test_user_active(self):
        '''
        Test users are active
        '''
        disabled_users = [
            x.get("email") for x in self.data if x.get("disabled")
        ]

        for u in self.final_config["users"]:
            self.assertTrue(
                u["mail"] not in disabled_users
            )

    def test_user_elligible(self):
        '''
        Test for elligible emails
        '''
        eligible_users_email = []
        for u in self.data:
            if "sre" in u.get("groups") or "tks" in u.get("groups"):
                eligible_users_email.append(u["email"])

        self.assertTrue(
            "sre@trustingsocial.com" in eligible_users_email
        )

        for u in self.final_config["users"]:
            print("Testing for user with email %s" % u.get("mail"))
            self.assertTrue(
                u["mail"] in eligible_users_email
            )

if __name__ == "__main__":
    unittest.main()
