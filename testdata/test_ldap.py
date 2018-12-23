#! /bin/python

import unittest

import os
import sys
import ldap


BIND = 'cn=ldap10,ou=sre,dc=tsocial,dc=com'
BIND_SECRET = "mysecret"

LDAP_HOST = os.environ.get("LDAP_HOST", 'ldap://localhost:8081')
SKIP_TLS = not LDAP_HOST.startswith("ldaps")

CWD = os.path.dirname(os.path.realpath(__file__))

class TestLDAPSearch(unittest.TestCase):
    SKIP_TLS = False

    def search(self, f):
        res = []
        ldap_attributes = ["*"]
        base_dn = 'dc=tsocial,dc=com'
        r_id = self.l.search(base_dn, ldap.SCOPE_SUBTREE, f, ldap_attributes)
        while True:
            r, d = self.l.result(r_id, 0)
            if r != ldap.RES_SEARCH_ENTRY:
                break

            res.append(r)
        return res

    @classmethod
    def setUpClass(cls):
        print(LDAP_HOST)
        if SKIP_TLS:
            return

        ldap.set_option(ldap.OPT_X_TLS_CACERTFILE,
                        os.path.join(CWD, 'cert-chain.pem'))

        ldap.set_option(ldap.OPT_X_TLS_CERTFILE,
                        os.path.join(CWD, 'client.pem'))

        ldap.set_option(ldap.OPT_X_TLS_KEYFILE,
                        os.path.join(CWD, 'client-key.pem'))

    def setUp(self):
        self.l = ldap.initialize(LDAP_HOST)
        self.l.protocol_version=ldap.VERSION3
        self.l.simple_bind_s(BIND, BIND_SECRET)

    def test_bad_password(self):
        with self.assertRaisesRegexp(Exception, "Invalid credentials"):
            self.l.simple_bind_s(BIND, "bad password")

    def tearDown(self):
        self.l.unbind()

    def test_search_common_name(self):
        res = self.search('cn=ldap20')
        self.assertEqual(len(res), 1, "record")

    def test_search_object_class(self):
        res = self.search("(objectClass=posixAccount)")
        self.assertEqual(len(res), 3, "Records")

    def test_search_uid(self):
        res = self.search("(uid=ldap20)")
        self.assertEqual(len(res), 1, "Records")

if __name__ == "__main__":
    unittest.main()

