module "edge-router" {
  source = "git@bitbucket.org:carepaydev/ssi-platform-modules.git//edge-router?ref=edge-router_1.0.5"

  country_code = "ke"
  stage        = "test"
  eks_id       = "3"

  ingress_dns_endpoint = "testserver.local"
}

module "another-router" {
  source = "git@bitbucket.org:carepaydev/ssi-platform-modules.git//edge-router"

  country_code = "ke"
  stage        = "test"
  eks_id       = "3"

  ingress_dns_endpoint = "testserver.local"
}
