---
marp: true
theme: nord
paginate: true
footer: "@borkod"
auto-scaling: true 
---

<!-- 

Title Slide

-->
# Automating Secure OIDC-Based Cross-Cloud Authentication

## Borko Djurkovic

---

<!--

Slide: Key Takeaways

-->
## Key Takeaways

- Do not use Access Keys and Secret Access Keys
- Use OIDC for secure system-to-system authentication
- Leverage automation at scale to improve security and reliability

---

<!--

Slide: About Me

Recommended podcast: Stuff You Should Know

Running: Rideau Canal in Ottawa

-->
## ABOUT ME

<style>
@import 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css';
</style>

![bg right](./images/ottawa.jpg)

👨🏻‍💻 Software Engineer

🗺 Ottawa, Canada

🏃🏻‍➡️ Runner

<i class="fa fa-podcast"></i> Podcast Listener

🔭 Physics Enthusiast

---


<!-- 

Slide: AWS IAM

- Creation and management of user identities (users, groups, and roles) within AWS.
- Each identity can be assigned **access keys** to authenticate to AWS resources.
- Supports multi-factor authentication (MFA) for additional security.

- Permissions are granted through policies that define what actions are allowed on specific resources.
- IAM roles used to assign permissions.
- Fine grained access control helps implement the principle of least privilege by ensuring users only have the permissions needed for their role.

- IAM integrates with external identity providers (e.g., Microsoft Entra ID, Okta) to enable Single Sign-On (SSO) capabilities.
- Supports both SAML (Security Assertion Markup Language) and OIDC (OpenID Connect) for federating access to AWS.
- Reducing the need for separate AWS credentials and improving security with centralized user management.

- AWS IAM enables issuance of temporary security credentials through AWS Security Token Service (STS).
- Ideal for situations requiring short-term access to AWS resources, such as third-party apps or external contractors.
- Temporary credentials are automatically revoked after a set time, reducing the risk of stale access permissions.

-->
## AWS IAM

### Provides Centralized Identity and Access Management

- Identity Management
  - Users, Groups, Roles
- RBAC and Permissions Management
  - IAM Roles, IAM Policies
- Federated Identities
  - SSO, SAML, OIDC
- AWS STS
  - Temporary Credentials and Session Management

---

<!--

Slide AWS IAM Authentication

- Access keys are used to authenticate users or services when interacting with AWS resources via the API, AWS CLI, or SDKs.
- Access Key ID is public
- Secret Access Key is private and must be kept secure.
- Together, they allow programmatic authentication without needing console credentials.
- The Access Key ID and Secret Access Key pair work like a username and password for API calls.
- Example: AWS Credentials File

-->
## Access Keys for Programmatic Authentication

- Access Key ID is public
- Secret Access Key is private
  - Must be kept secure

```bash
[user1]
aws_access_key_id=ASIAI44QH8DHBEXAMPLE
aws_secret_access_key=je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
```

---

<!--

Slide: DO NOT USE SECRET KEYS

- Long-term access keys can be compromised if exposed or mishandled, leading to persistent security risks.
- Against prescribed AWS security best practices.
- Managing access keys across multiple users, services, and systems becomes complex as environments grow.
- Access keys are harder to track and audit compared to roles with temporary credentials.
- Against corporate, enterprise, or regulatory security policies.
- Access keys are tied to specific IAM users, not roles or fine-grained policies.

-->
## DO NOT USE ACCESS KEYS

- Security risk with long term credentials
- Against security best practices
- Difficult to manage at scale
- Against corporate security policies
- Lack of granular access control

---

<!--

Slide: OIDC

- OIDC is an authentication layer built on top of OAuth 2.0, enabling identity verification using tokens.
- Additionally, these tokens that carry claims (user info) and scopes (permissions).
- This allows fine-grained authorization in apps by controlling what data users can access based on their identity.
- It allows applications to authenticate users via external identity providers (e.g., Google, Microsoft).
- OIDC is commonly used for Single Sign-On (SSO), allowing users to log in once and access multiple apps.
- Reduces the need for separate credentials.
- OIDC provides more secure and scalable user authentication compared to long-term AWS IAM access keys, which are prone to exposure and management overhead.
- OIDC tokens are short-lived and can be dynamically scoped, reducing the risk of stale or overly broad permissions that are common with IAM access keys.

-->
## Better Alternative: Use OIDC

### OIDC enables federated identity management across applications and services

- Provides identity verification using tokens
- Enables fine-grained authorization using claims and scopes
- OIDC tokens are short lived and dynamically scoped
- Enables federated access allowing IdP's to authenticate users for third party apps

---

<!--

Slide: AWS OIDC

- AWS IAM OIDC (OpenID Connect) allows external identity providers (like Google or Microsoft Entra ID) to authenticate users for AWS services.
- It enables federated access, allowing users to sign in using their existing credentials, without needing separate AWS IAM user accounts.
- Supports temporary short-lived credentials
- Ideal for organizations that want to authenticate users from external identity providers like Google, Facebook, or enterprise SSO solutions.
- Useful for scenarios where users need access to AWS resources but should not have IAM user credentials (e.g., external contractors, third-party services).
- Reduces the risk of credential leakage by using short-lived tokens instead of long-term AWS IAM access keys.
- IAM Web Identity Roles allow users to assume AWS roles using tokens from external identity providers (such as OIDC-compatible services like Google, Facebook, or custom enterprise SSO solutions).

-->
## AWS IAM OIDC Provider

- Integration with external OIDC IdP's for authentication to AWS
- Enables federated access and SSO
- Removes need for AWS IAM users and long-lived AWS IAM access keys
- External users assume AWS IAM Web Identity Roles

---

<!--

Slide: ENTRA ID

- Entra ID (formerly Azure Active Directory) allows organizations to manage users.
- Comprehensive cloud-based IdP
- Provides tools to manage lifecycle of users such as provisioning and deprovisioning, access permissions.
- Governance and security - conditional access policies, MFA, advanced reporting
- Federated access enabling SSO via SAML or OIDC
- Provides ability to register applications and configuring OIDC, enabling obtaining tokens for API access

-->
## Entra ID = Azure based IdP

- Microsoft Azure based IAM solution
- Provides IdP functionality (e.g. user management)
- Application registration for external apps
- Identity governance

---

<!--

Slide: Integrated Solution

- An external application (e.g. CI/CD) is registered in Entra ID.
- An external application logs in with an Azure service principal.
- When a login with the service principal is successful, it authorizes the use of the Azure App Registration and provides the client with a token that specifies the roles, audience, and permissions available to the client in Entra ID.
- The client uses the token obtained from Microsoft Entra ID to exchange it with a temporary token from AWS STS.
- AWS STS generates a temporary token and provides it to the client.
- The client is able to take on the IAM Web Identity Role and gain access to the resources permitted for that role.
- A trust policy within the IAM Web Identity Role authorizes AWS STS to exchange an Entra ID token for an AWS STS token.

-->
## AWS + Entra ID Integrated Solution

![bg fit right](./images/integrated-solution-dark.png)

1. External application is registered in Entra ID
2. Upon login, Entra ID provides a token
3. The client uses the Entra ID token to obtain a temporary token from AWS STS
4. The client assumes the IAM Web Identity Role

---

<!--

Slide: Manual setup - App Registration

-->
## Manual Configuration - Application Registration in Entra ID

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:550 center](./images/entra-app-registration-manual.png)

---

<!--

Slide: Manual setup - App Registration

-->
## Manual Configuration - Application Registration in Entra ID

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:550 center](./images/entra-app-registration-manual-2.png)

---

<!--

Slide: Manual setup - AWS IAM OIDC Provider Configuration

OIDC Provider configured for the Entra ID tenant

-->
## Manual Configuration - AWS IAM OIDC Provider Configuration

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:300 center](./images/aws-oidc-aud-manual.png)

---

<!--

Slide: Manual setup - IAM Web Identity Role Configuration

-->
## Manual Configuration - IAM Web Identity Role Configuration

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:500 center](./images/aws-iam-create-web-role-manual.png)

---

<!--

Slide: Manual setup - IAM Web Identity Role Configuration

-->
## Manual Configuration - IAM Web Identity Role Configuration

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

![bg fit right](./images/aws-iam-create-web-role-manual-2.png)

---

<!--

Slide 10: Challenges of Manual Setup in Enterprise Environments

- Results in a violation of enterprise security policies and compliance because, when done by different individuals, it will result in varying approaches and setups.
- Needs substantial access privileges in both AWS and the IdP, potentially leading to significant security risks.
- Requires a deep understanding of both platforms, a competence that can be lacking in many companies.


-->
## Challenges of Manual Setup in Enterprise Environments

- Violation of enterprise security policies
- Requires elevated privileges in both AWS and Entra ID
- Requires deep understanding of both platforms (AWS IAM and Entra ID)
- Different individuals use varying approaches and setups
- Manual configurations are error prone

---

<!--

Slide: Need for automation at scale

- Enterprise organizations multiple AWS accounts for various purposes (e.g. Logging, Networking, Security).
- Additionally, various business units may have many AWS Accounts.
- Environments (e.g. DEV, TEST, PROD) are often broken up into different AWS accounts.
- Many organizations uses AWS Control Tower, AWS Landing Zone Accelerator (LZA) or other automation pipelines to programmatically create AWS accounts at scale.
- In these scenarios, automation can create AWS IAM OIDC Provider in each account.

-->
## Need for automation at scale

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:400 center](./images/lza.png)

---

<!--

Slide: Solution overview

- Process starts with creation or deletion of an IAM Web Identity Role in one of the newly created member accounts.
- AWS CloudTrail records the event.
- Event Rule captures these specific events and sends them to an Event Bus
- In the OIDC Factory Account, we also have an event rule
- Event rule triggers a Lambda Function that then invokes one of two Step Function workflows.

-->
## Solution overview

<style>
img[alt~="center"] {
  display: block;
  margin: 0 auto;
}
</style>

![height:600 center](./images/automation-architecture-dark.png)

---

<!--

Slide: Solution Components - Creation of IAM Web Identity Role

- **Create Service Principal Lambda**
  - Function calls a Entra ID API to register a new application. The Lambda function returns a unique audience identifier for the Entra ID application.

- **Add Audience to OIDC Provider Lambda**
  - Function receives an audience identifier. It adds the audience identifier to the pre-deployed IAM OIDC Provider in the newly-created AWS account.

- **Assign Role to Audience Lambda**
  - Function updates the trust relationship in the IAM Web Identity Role by adding the audience identifier of the application registration to allow the newly-created service principal to assume the Web Identity Role.
  - This step is required because, unlike a manual setup, the AWS user does not have an audience identifier so users enter any dummy value which will be overwritten few seconds afterwards.

-->
## Creation of IAM Web Identity Role

![bg fit right](./images/automation-architecture-dark.png)

<style scoped>
section {
    font-size: 22px;
}
</style>

- **Create Service Principal Lambda**
  - Invokes Entra ID API to register a new application.

- **Add Audience to OIDC Provider Lambda**
  - Adds the audience identifier to the IAM OIDC Provider.

- **Assign Role to Audience Lambda**
  - Updates the trust relationship in the IAM Web Identity Role to add the audience identifier and enable the service principal to assume the role.

---

<!--

Slide: Solution Components - Deletion of IAM Web Identity Role

- **Delete Service Principal Lambda**
  - Invokes Entra ID API to delete the existing service principal.

- **Remove Audience from OIDC Provider Lambda**
  - Removes the audience identifier from the IAM OIDC Provider.

-->
## Deletion of IAM Web Identity Role

![bg fit right](./images/automation-architecture-dark.png)

<style scoped>
section {
    font-size: 22px;
}
</style>

- **Delete Service Principal Lambda**
  - Invokes Entra ID API to delete the existing service principal.

- **Remove Audience from OIDC Provider Lambda**
  - Removes the audience identifier from the IAM OIDC Provider.

---

<!-- Slide: Key Takeaways -->
## Key Takeaways

- Do not use Access Keys and Secret Access Keys
- Use OIDC for secure system-to-system authentication
- Leverage automation at scale to improve security and reliability

---

<!-- Slide: Info & Source Code -->
## Additional Info

<style>
@import 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css';
</style>

<i class="fa fa-github"></i> https://github.com/borkod/aws-oidc-automation

<i class="fa fa-globe"></i> https://www.b3o.tech/

<svg xmlns="http://www.w3.org/2000/svg" height="32" width="32" viewBox="0 0 640 640"><path fill="#ffffff" d="M439.8 358.7C436.5 358.3 433.1 357.9 429.8 357.4C433.2 357.8 436.5 358.3 439.8 358.7zM320 291.1C293.9 240.4 222.9 145.9 156.9 99.3C93.6 54.6 69.5 62.3 53.6 69.5C35.3 77.8 32 105.9 32 122.4C32 138.9 41.1 258 47 277.9C66.5 343.6 136.1 365.8 200.2 358.6C203.5 358.1 206.8 357.7 210.2 357.2C206.9 357.7 203.6 358.2 200.2 358.6C106.3 372.6 22.9 406.8 132.3 528.5C252.6 653.1 297.1 501.8 320 425.1C342.9 501.8 369.2 647.6 505.6 528.5C608 425.1 533.7 372.5 439.8 358.6C436.5 358.2 433.1 357.8 429.8 357.3C433.2 357.7 436.5 358.2 439.8 358.6C503.9 365.7 573.4 343.5 593 277.9C598.9 258 608 139 608 122.4C608 105.8 604.7 77.7 586.4 69.5C570.6 62.4 546.4 54.6 483.2 99.3C417.1 145.9 346.1 240.4 320 291.1z"/></svg> [@b3o.tech](https://bsky.app/profile/b3o.tech)

<svg xmlns="http://www.w3.org/2000/svg" height="32" width="32" viewBox="0 0 640 640"><path fill="#ffffff" d="M453.2 112L523.8 112L369.6 288.2L551 528L409 528L297.7 382.6L170.5 528L99.8 528L264.7 339.5L90.8 112L236.4 112L336.9 244.9L453.2 112zM428.4 485.8L467.5 485.8L215.1 152L173.1 152L428.4 485.8z"/></svg> [@borkod](https://x.com/borkod)

<i class="fa fa-linkedin"></i> https://www.linkedin.com/in/borkod/

---
