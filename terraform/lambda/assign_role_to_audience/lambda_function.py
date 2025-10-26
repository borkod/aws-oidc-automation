import json
import logging
import boto3
import os
from botocore.exceptions import ClientError

logger = logging.getLogger()
logger.setLevel(logging.INFO)

OIDC_PROVIDER_ARN_TEMPLATE = "arn:aws:iam::{account_id}:oidc-provider/{oidc_url}"
CROSS_ACCOUNT_ROLE_ARN_TEMPLATE = "arn:aws:iam::{account_id}:role/{cross_account_role_name}"

def assume_role(account_id, role_name):
    sts_client = boto3.client("sts")
    role_arn = CROSS_ACCOUNT_ROLE_ARN_TEMPLATE.format(account_id=account_id, cross_account_role_name=role_name)
    try:
        response = sts_client.assume_role(
            RoleArn=role_arn,
            RoleSessionName="UpdateTrustRelationshipSession"
        )
        credentials = response["Credentials"]
        return credentials
    except ClientError as e:
        logger.error(f"Error assuming role {role_arn}: {e}")
        raise

def update_trust_relationship(iam_client, account_id, role_name, audience, oidc_url):
    try:
        role = iam_client.get_role(RoleName=role_name)
        trust_policy = role["Role"]["AssumeRolePolicyDocument"]

        updated = False
        oidc_provider_arn = OIDC_PROVIDER_ARN_TEMPLATE.format(account_id=account_id, oidc_url=oidc_url)
        audience_key = f"{oidc_url}:aud"
        audience = f"api://{audience}"

        for stmt in trust_policy.get("Statement", []):
            if (
                stmt.get("Principal", {}).get("Federated") == oidc_provider_arn and
                stmt.get("Action") == "sts:AssumeRoleWithWebIdentity" and
                "Condition" in stmt and
                "StringEquals" in stmt["Condition"] and
                audience_key in stmt["Condition"]["StringEquals"]
            ):
                stmt["Condition"]["StringEquals"][audience_key] = [audience]
                updated = True

        if not updated:
            logger.error("OIDC trust relationship statement not found or malformed.")
            raise Exception("OIDC trust relationship statement not found or malformed.")

        iam_client.update_assume_role_policy(
            RoleName=role_name,
            PolicyDocument=json.dumps(trust_policy)
        )
        logger.info(f"Updated trust relationship for role {role_name} in account {account_id} with audience {audience}")
        return True
    except ClientError as e:
        logger.error(f"Error updating trust relationship for role {role_name}: {e}")
        raise

def lambda_handler(event, context):
    logger.info(f"Received event: {json.dumps(event)}")

    sfn_param = event.get("sfnParam", {})
    account_id = sfn_param.get("account")
    role_name = sfn_param.get("roleName")
    audience = event.get("audience")
    oidc_url = os.environ.get("OIDC_URL")
    cross_account_role_name = os.environ.get("CROSS_ACCOUNT_ROLE_NAME")

    if not account_id or not role_name or not audience or not oidc_url or not role_name:
        logger.error("Missing required parameters.")
        raise Exception("Missing required parameters.")

    # Assume role in target account
    credentials = assume_role(account_id, cross_account_role_name)

    iam_client = boto3.client(
        "iam",
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )

    update_trust_relationship(iam_client, account_id, role_name, audience, oidc_url)

    return {
        "status": "success",
        "roleName": role_name,
        "account": account_id,
        "audience": audience
    }