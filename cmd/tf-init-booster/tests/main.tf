module "edge-router" {
  source = "git@bitbucket.org:carepaydev/ssi-platform-modules.git//edge-router?ref=edge-router_2.4.0"

  country_code = "ke"
  stage        = "test"
  eks_id       = "3"

  ingress_dns_endpoint = "testserver.local"
}
