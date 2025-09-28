# aws-oidc-automation

This repository accompanies the presentation at the Commit Your Code conference. The presentation can be found [here](https://github.com/borkod/CYC2025).

It includes Terraform and Lambda source code to set up all resources required for the OIDC automation account configuration.

## Problem Description

Imagine you have an external application (let's say a CI/CD system) that needs to interact with your AWS account. This could be for deploying applications, updating resources, or provisioning new infrastructure. No big deal, right?

But here's the kicker: this CI/CD system should run on autopilot, with minimal user input. It might kick off pipelines in the early morning hours based on a schedule, trigger a deployment when a PR is merged, or even run in the middle of the night to respond to an incident.

So, how do you ensure the system can authenticate to AWS to carry out these tasks?

### Enter IAM Users...

A quick and straightforward solution would be to create an IAM user for the CI/CD system, generate an Access Key ID and Secret Access Key, and let it run free. That’s one way to get the job done.

But here’s the catch: Access Key ID and Secret Access Key are essentially a static username/password combo. They’re long-lived credentials; meaning they don’t expire unless you manually rotate them, and they’re a major security risk if exposed.

In other words, if these credentials are ever leaked, your AWS environment is wide open to exploitation. And that’s not something you want, especially for a system running unattended.

There are several other issues with relying on Access Key IDs for AWS authentication:
- They are against AWS best security practices and recommendations
- In large enterprise environment, they are likely to be in contradiction with corporate security policies
- In organizations that must comply with regulatory policies (e.g. HIPAA), use of long-lived credentials is against those regulatory security policies. 

- **They go against AWS security best practices**: AWS strongly recommends using temporary credentials (like those from IAM roles) instead of long-lived access keys.
- **They often conflict with enterprise security policies**: In large organizations, permanent credentials can violate internal compliance standards and raise red flags with security teams.
- **They may breach regulatory compliance**: For environments governed by regulations like HIPAA, the use of long-lived credentials can directly contradict security requirements and expose the organization to compliance risks.

### The Quick Fix: Rotate Credentials...

Many organizations go down the road of building Frankenstein-like scripts to rotate Access Key ID/Secret Access Keys regularly. While this can help mitigate the risk of exposed credentials, it’s more of a band-aid solution than a sustainable fix.

So, what's the better way to handle this situation, you ask?

## OIDC

OpenID Connect (OIDC) is an authentication protocol built on top of OAuth 2.0. It enables applications to verify users’ identities using tokens issued by an external identity provider (IdP). These tokens contain valuable information, such as claims (user attributes) and scopes (permissions), allowing for more contextual and secure access control.

One of the key advantages of OIDC is its support for **short-lived, dynamically scoped** tokens. Unlike long-lived AWS IAM access keys, OIDC tokens are designed to expire quickly and can be fine-tuned for specific permissions. This significantly reduces the risk of overly broad access and stale credentials, improving both security and manageability.

OIDC also powers Single Sign-On (SSO), allowing users to authenticate once and access multiple applications without re-entering credentials. This is especially valuable in enterprise environments where user experience and centralized access control are critical. With OIDC, applications can offload authentication to trusted identity providers like Microsoft Entra ID (formerly Azure Active Directory), eliminating the need to manage separate AWS IAM credentials for every user.

### Secure Federated Access with AWS and OIDC

AWS supports federated authentication through IAM OIDC Providers, which allow external identity providers to securely authenticate users and grant them access to AWS services. This approach is particularly beneficial in scenarios where users (e.g. external contractors, third-party apps) require temporary access to AWS resources but should not have long-term IAM user credentials.

By using IAM Web Identity Roles, AWS can issue temporary credentials to authenticated users based on policies attached to the role. This eliminates the need for creating and managing dedicated IAM users, and it drastically reduces the risk of credential leakage.

### Microsoft Entra ID as an OIDC Identity Provider

Microsoft Entra ID is a powerful, cloud-based identity platform that provides robust tools for user lifecycle management. These include user provisioning, deprovisioning, and access control. It is frequently used at large organizations for managing enterprise-scale identities.

Entra ID also supports advanced governance and security features such as Conditional Access policies, Multi-Factor Authentication (MFA), and rich auditing and reporting capabilities. It can federate access to AWS through both OIDC and SAML, enabling organizations to centralize identity and access management while maintaining strong security posture.

Furthermore, Entra ID supports the registration of external applications, allowing developers to configure OIDC clients for secure, standards-based authentication across environments.

## Issues with Manual Setup

### Challenges of Manual Setup in Enterprise Environments

Manually configuring identity federation between AWS and Microsoft Entra ID can introduce several risks in enterprise settings.

- manual setup often leads to inconsistencies, especially when multiple individuals are involved. Each person may take a slightly different approach, resulting in misaligned configurations that can violate enterprise security policies and compliance requirements. In regulated industries, this lack of standardization can have serious consequences.
- the process requires elevated access privileges across both AWS and Entra ID. Granting broad permissions for manual tasks increases the risk of accidental misconfigurations or, worse, exposure of sensitive resources—creating a significant security risk.
- Another common hurdle is the depth of expertise required. Setting up a secure and reliable federation between these platforms demands a strong understanding of both AWS IAM and Entra ID’s identity and access management features. Many organizations struggle with this due to limited in-house expertise or competing priorities within their IT teams.
- Finally, it’s important to recognize that manual configuration is inherently error-prone. A missed setting, incorrect permission, or misapplied policy can lead to hours of troubleshooting—or a costly security gap.

For these reasons, automation and standardized tooling are strongly recommended when integrating identity providers in enterprise-scale cloud environments.

### The Need for Automation at Scale

In enterprise environments, it's common to manage dozens—or even hundreds—of AWS accounts. These accounts are often segmented by function—such as logging, networking, or security—and distributed across various business units, each maintaining its own set of accounts.

To further complicate things, most organizations separate development, testing, and production environments into individual AWS accounts for better isolation and governance. Temporary or short-lived accounts are also frequently created for use cases like developer sandboxes, proof-of-concepts (POCs), or training labs.

To manage this complexity, many organizations turn to solutions like AWS Control Tower, the AWS Landing Zone Accelerator (LZA), or custom-built automation pipelines. These tools allow for the programmatic creation of AWS accounts at scale, helping teams maintain consistency and control across the organization.

However, simply creating accounts isn’t enough. As identity federation becomes the standard for secure access, there’s also a need to automate the setup of AWS IAM OIDC Providers in each account. Manual configuration is impractical at this scale—automation is essential to ensure that OIDC configurations are applied consistently, securely, and efficiently across all accounts.