import os
import json
import logging
import boto3
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
            RoleSessionName="AddOIDCAudience"
        )
        credentials = response["Credentials"]
        return credentials
    except ClientError as e:
        logger.error(f"Error assuming role {role_arn}: {e}")
        raise

def delete_audience(iam_client, account_id, audience, oidc_url):
    oidc_provider_arn = OIDC_PROVIDER_ARN_TEMPLATE.format(account_id=account_id, oidc_url=oidc_url)
    try:
        response = iam_client.remove_client_id_from_open_id_connect_provider(
            OpenIDConnectProviderArn=oidc_provider_arn,
            ClientID=audience
        )

        return True
    except ClientError as e:
        logger.error(f"Error updating OIDC provider: {e}")
        raise

def lambda_handler(event, context):
    logger.info(f"Received event: {json.dumps(event)}")

    sfn_param = event.get("sfnParam", {})
    account_id = sfn_param.get("account")
    audience = event.get("audience")
    audience = f"api://{audience}"
    oidc_url = os.environ.get("OIDC_URL")
    role_name = os.environ.get("CROSS_ACCOUNT_ROLE_NAME")

    if not account_id or not audience or not oidc_url:
        logger.error("Missing required parameters")
        raise Exception("Missing required parameters.")

    # Assume role in target account
    credentials = assume_role(account_id, role_name)

    iam_client = boto3.client(
        "iam",
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )

    delete_audience(iam_client, account_id, audience, oidc_url)

    return {
            "status": "success",
            "account": account_id,
            "oidc_url": oidc_url,
            "audience": audience
        }