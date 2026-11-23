import importlib.util,pathlib,unittest,tempfile,os
spec=importlib.util.spec_from_file_location('refresh',pathlib.Path(__file__).with_name('refresh.py'))
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
class InventoryTests(unittest.TestCase):
 def test_literal_quote_and_delimiter(self):
  self.assertIn("a'b",m.lit("a'b"))
  with self.assertRaises(AssertionError):m.lit('$inventory$')
 def test_missing_disabled_and_role_scope(self):
  rows=m.make_rows([dict(name='nmap',purpose='network',path=None,ready=False)],['planner','worker','reporter'])
  self.assertEqual(rows[0]['agents'],['planner','worker','reporter'])
  self.assertEqual(rows[1]['agents'],['worker']);self.assertFalse(rows[1]['enabled'])
 def test_present_is_not_full_readiness(self):
  row=m.make_rows([dict(name='jq',purpose='json',path='/usr/bin/jq',ready=True)],['worker'])[1]
  self.assertTrue(row['enabled']);self.assertIn('functionality beyond that is not proven',row['description'])
 def test_snapshot_readable_under_restrictive_umask(self):
  with tempfile.TemporaryDirectory() as d:
   p=pathlib.Path(d)/'inventory.md';old=os.umask(0o077)
   try:m.write_file(p,'complete snapshot')
   finally:os.umask(old)
   self.assertEqual(p.stat().st_mode & 0o777,0o644)
   self.assertEqual(p.read_text(),'complete snapshot')
   m.write_file(p,'private state',0o600)
   self.assertEqual(p.stat().st_mode & 0o777,0o600)
if __name__=='__main__':unittest.main()
