# aws-oidc-automation

This repository accompanies the presentation at the Commit Your Code conference. The presentation can be found [here](https://github.com/borkod/CYC2025).

It includes Terraform and Lambda source code to set up all resources required for the OIDC automation account configuration.

## Problem Description

Imagine you have an external application (let's say a CI/CD system) that needs to interact with your AWS account. This could be for deploying applications, updating resources, or provisioning new infrastructure. 

This CI/CD system should also be able run on autopilot, with minimal user input. It might kick off pipelines in the early morning hours based on a schedule, trigger a deployment when a PR is merged, or even run in the middle of the night to respond to an incident.

So, how do you ensure the system can authenticate to AWS to carry out these tasks?

### Enter IAM Users...

A quick and straightforward solution would be to create an IAM user for the CI/CD system, generate an Access Key ID and Secret Access Key. That’s one way to get the job done.

But here’s the catch: Access Key ID and Secret Access Key are essentially a static username/password combo. They’re long-lived credentials; meaning they don’t expire unless you manually rotate them, and they’re a major security risk if exposed.

In other words, if these credentials are ever leaked, your AWS environment is wide open to exploitation. And that’s not something you want, especially for a system running unattended.

There are several other issues with relying on Access Key IDs for AWS authentication:
- **They go against AWS security best practices**: AWS strongly recommends using temporary credentials (like those from IAM roles) instead of long-lived access keys.
- **They often conflict with enterprise security policies**: In large organizations, permanent credentials can violate internal compliance standards and raise red flags with security teams.
- **They may breach regulatory compliance**: For environments governed by regulations like HIPAA, the use of long-lived credentials can directly contradict security requirements and expose the organization to compliance risks.

### The Quick Fix: Rotate Credentials...

Many organizations go down the road of building Frankenstein-like scripts to rotate Access Key ID/Secret Access Keys regularly. While this can help mitigate the risk of exposed credentials, it’s more of a band-aid solution than a sustainable fix.

So, what's the better way to handle this situation?

## OIDC

OpenID Connect (OIDC) is an authentication protocol built on top of OAuth 2.0. It enables applications to verify users’ identities using tokens issued by an external identity provider (IdP). These tokens contain valuable information, such as claims (user attributes) and scopes (permissions), allowing for more contextual and secure access control.

One of the key advantages of OIDC is its support for **short-lived, dynamically scoped** tokens. Unlike long-lived AWS IAM access keys, OIDC tokens are designed to expire quickly and can be fine-tuned for specific permissions. This significantly reduces the risk of overly broad access and stale credentials, improving both security and manageability.

OIDC also enables Single Sign-On (SSO), allowing users to authenticate once and access multiple applications without re-entering credentials. This is especially valuable in enterprise environments where centralized access control are critical. With OIDC, applications can offload authentication to trusted identity providers like Microsoft Entra ID (formerly Azure Active Directory), eliminating the need to manage separate AWS IAM credentials for every user.

### Secure Federated Access with AWS and OIDC

AWS supports federated authentication through IAM OIDC Providers, which allow external identity providers to securely authenticate users and grant them access to AWS services. This approach is particularly beneficial in scenarios where users (e.g. external contractors, third-party apps) require temporary access to AWS resources but should not have long-term IAM user credentials.

By using IAM Web Identity Roles, AWS can issue temporary credentials to authenticated users based on policies attached to the role. This eliminates the need for creating and managing dedicated IAM users, and it drastically reduces the risk of credential leakage.

### Microsoft Entra ID as an OIDC Identity Provider

Microsoft Entra ID is a powerful, cloud-based identity platform that provides robust tools for user lifecycle management. These include user provisioning, deprovisioning, and access control. It is frequently used at large organizations for managing enterprise-scale identities.

Entra ID also supports advanced governance and security features such as Conditional Access policies, Multi-Factor Authentication (MFA), and rich auditing and reporting capabilities. It can federate access to AWS through both OIDC and SAML, enabling organizations to centralize identity and access management while maintaining strong security posture.

Furthermore, Entra ID supports the registration of external applications, allowing developers to configure OIDC clients for secure, standards-based authentication across environments.

### Integrated Solution

![Integrated Solution](./images/integrated-solution-dark.png)

To enable secure, federated access from external systems such as CI/CD pipelines, an external application is first registered in Microsoft Entra ID. This application authenticates using an Azure service principal, which acts as its identity. Upon successful authentication, Entra ID issues an OIDC token to the application. This token includes important details such as the audience, roles, and permissions the application is allowed to use within Entra ID.

The application then uses this Entra ID-issued token to request temporary AWS credentials from AWS Security Token Service (STS). If the request is valid and the IAM Web Identity Role's trust policy authorizes Entra ID as a trusted identity provider, AWS STS exchanges the OIDC token for a temporary AWS token. With these credentials, the external application can assume the IAM role and gain access to the AWS resources defined in that role's permissions. The entire flow is done securely and without the need for long-term IAM credentials.

## Manual Configuration

Manual configuration of IAM and Entra ID integration consists of three steps:

1. Create Application Registration in Entra ID
2. Add Audience to AWS IAM OIDC Provider
3. Create AWS IAM Web Identity Role

### Pre-requisite: IAM OICD Provider

TODO: add steps for this

### Entra ID Application Registration

TODO: Confirm steps

In Entra ID, under the Applications tab, create a new Application Registration:

![Entra ID App Registration](./images/manual/entra-app-registration-manual.png)

Once created, Entra ID will provide an Application / Client ID:

![Entra ID App Registration Client ID](./images/manual/entra-app-registration-manual-2.png)

### Add Audience to the AWS IAM OIDC Provider

Using the Application / Client ID value obtained from Entra ID, we add a new audience value to the Entra ID IAM OIDC provider:

![AWS IAM OIDC Provider Audience](./images/manual/aws-oidc-aud-manual.png)

### AWS IAM Web Identity Role

Final step is to create AWS IAM Web Identity Role. We specify the Entra ID IAM OIDC Provider, and our application registratration audience value:

![AWS IAM Web Identity Role](./images/manual/aws-iam-create-web-role-manual.png)

Below shows the IAM trust policy that enables the integration from IAM Web Identity Role to the Entra ID application registration:

```json
{
    "Effect": "Allow",
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Principal": {
        "Federated":
        "arn:aws:iam::<account>:oidc-provider/
        sts.windows.net/<tenant>"
    },
    "Condition": {
        "StringEquals": {
            "sts.windows.net/tenant:aud": [
                "<audience>"
            ]
        }
    }
}
```

![AWS IAM Web Identity Role Trust Policy](./images/manual/aws-iam-create-web-role-manual-2.png)

## Issues with Manual Setup

### Challenges of Manual Setup in Enterprise Environments

Manually configuring identity federation between AWS and Microsoft Entra ID can introduce several risks in enterprise settings:

Manual setup often leads to inconsistencies, especially when multiple individuals are involved. Each person may take a slightly different approach, resulting in misaligned configurations that can violate enterprise security policies and compliance requirements. In regulated industries, this lack of standardization can have serious consequences.

The process requires elevated access privileges across both AWS and Entra ID. Granting broad permissions for manual tasks increases the risk of accidental misconfigurations or, worse, exposure of sensitive resources—creating a significant security risk.

Another common hurdle is the depth of expertise required. Setting up a secure and reliable federation between these platforms demands a strong understanding of both AWS IAM and Entra ID’s identity and access management features. Many organizations struggle with this due to limited in-house expertise or competing priorities within their IT teams.

Finally, it’s important to recognize that manual configuration is inherently error-prone. A missed setting, incorrect permission, or misapplied policy can lead to hours of troubleshooting—or a costly security gap.

For these reasons, automation and standardized tooling are strongly recommended when integrating identity providers in enterprise-scale cloud environments.

### The Need for Automation at Scale

In enterprise environments, it's common to manage dozens—or even hundreds—of AWS accounts. These accounts are often segmented by function—such as logging, networking, or security—and distributed across various business units, each maintaining its own set of accounts.

To further complicate things, most organizations separate development, testing, and production environments into individual AWS accounts for better isolation and governance. Temporary or short-lived accounts are also frequently created for use cases like developer sandboxes, proof-of-concepts (POCs), or training labs.

To manage this complexity, many organizations turn to solutions like AWS Control Tower, the AWS Landing Zone Accelerator (LZA), or custom-built automation pipelines. These tools allow for the programmatic creation of AWS accounts at scale, helping teams maintain consistency and control across the organization.

However, simply creating accounts isn’t enough. As identity federation becomes the standard for secure access, there’s also a need to automate the setup of AWS IAM OIDC Providers in each account. Manual configuration is impractical at this scale—automation is essential to ensure that OIDC configurations are applied consistently, securely, and efficiently across all accounts.

![Automation at scale](./images/lza.png)

AWS Landing Zone IAM Config does not have built-in support for configuring IAM OIDC Providers. So instead we have to create our own Cloud Formation Stack to provision the Entra ID OIDC Provider and add it to the Landing Zone Customizations.

Links:

https://awslabs.github.io/landing-zone-accelerator-on-aws/latest/user-guide/config/#customization-configuration

https://awslabs.github.io/landing-zone-accelerator-on-aws/latest/typedocs/interfaces/___packages__aws_accelerator_config_lib_models_customizations_config.ICloudFormationStack.html


Sample code in `customizations-config.yaml`:

```yaml
customizations:
  cloudFormationStacks:
    - name: EntraIDOIDCProvider
      description: Deploy EntraID OIDC Provider to Workload accounts
      deploymentTargets:
        organizationalUnits:
            - ClientWorkloads
      runOrder: 1
      regions:
        - *HOME_REGION
      template: cloudformation/EntraIDOIDCProviderStack.yaml
...
```

Sample code in `cloudformation/EntraIDOIDCProviderStack.yaml`:

```yaml
Resources:
  EntraIDOIDCProvider:
    Type: AWS::IAM::OIDCProvider
    Properties:
      Url: !Sub "https://${EntraIDURL}"
      ClientIdList:
        - !Sub "https://${EntraIDURL}"
      ThumbprintList:
        - !Ref EntraIDThumbprint
....
```
https://awslabs.github.io/landing-zone-accelerator-on-aws/latest/typedocs/interfaces/___packages__aws_accelerator_config_lib_models_customizations_config.ICloudFormationStack.html

https://github.com/awslabs/landing-zone-accelerator-on-aws/issues/74#issuecomment-1515659109

## OIDC Configuration - Automated Solution

On the left, we have a sample account (Account A) which is a newly created account (created via automation).

On the right, we have a centralized OIDC factory account that implements the automation solution.

Process starts with creation or deletion of an IAM Web Identity Role in one of the newly created member accounts. AWS CloudTrail records this event.An Event Bus Event Rule captures these specific events and sends them to an Event Bus.

TODO: Code for the event rule

```json
code
```

In the OIDC Factory Account, we also have an event rule.

TODO: Add how they are connected.

Event rule triggers a Lambda Function that then invokes one of two Step Function workflows. First workflow deals with events related to the creation of IAM Web Identity Roles.  The second workflow deals with the deletion of IAM Web Identity Roles.

### Solution Components - Creation of IAM Web Identity Role

The first workflow consists of three Lambda functions:

- **Create App Registration Lambda**
  - Function calls a Entra ID API to register a new application.
  - The Lambda function returns a unique audience identifier for the Entra ID application.

- **Add Audience to OIDC Provider Lambda**
  - Function receives an audience identifier. It adds the audience identifier to the pre-deployed IAM OIDC Provider in the newly-created AWS account (Account A).

- **Assign Role to Audience Lambda**
  - Function updates the trust relationship in the IAM Web Identity Role by adding the audience identifier of the application registration.
  - This step is required because, unlike a manual setup, the AWS user that creates the web identity role does not have an audience identifier so users enter any dummy value. This will be overwritten few seconds later by the automation.

TODO: Update image to include Entra ID, IAM OIDC Provider, IAM Web Identity Role
![Workflow 1](./images/automation-workflow-1-dark.png)

  ### Solution Components - Deletion of IAM Web Identity Role

- **Delete Service Principal Lambda**
  - Invokes Entra ID API to delete the existing app registration.

- **Remove Audience from OIDC Provider Lambda**
  - Removes the audience identifier from the IAM OIDC Provider.

TODO: Update image to include Entra ID, OAIM OIDC Provider
![Workflow 2](./images/automation-workflow-2-dark.png)

## Summary

To recap, the key takeaways are:
- We should not be using Access Key ID/Secret Access Key as a method of authentication into AWS in any Production environments or environments that require higher degree of security.
- A better approach is to use the AWS IAM OIDC Provider to integrate with an external Identity Provider for syste-to-system authetication.
- In complex and large scale environments, automation is useful to improve security and reliability of OIDC configurations.