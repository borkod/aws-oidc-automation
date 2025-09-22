import os
import json
import logging
import boto3
from botocore.exceptions import ClientError

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Step Functions client
sfn_client = boto3.client('stepfunctions')

# Environment variables (Step Function ARNs)
CREATE_ROLE_SFN_ARN = os.getenv("CREATE_ROLE_SFN_ARN")
DELETE_ROLE_SFN_ARN = os.getenv("DELETE_ROLE_SFN_ARN")

def lambda_handler(event, context):
    """
    Lambda handler to process IAM role events and invoke the corresponding Step Function.
    """
    try:
        # Extract fields safely
        account_number = event.get('account')
        event_name = event.get('detail', {}).get('eventName')
        role_name = event.get('detail', {}).get('requestParameters', {}).get('roleName')

        logger.info(f"Received event for account: {account_number}, event: {event_name}, role: {role_name}")

        if event_name == "CreateRole":
            return start_step_function(CREATE_ROLE_SFN_ARN, account_number, event_name, role_name)

        elif event_name == "DeleteRole":
            return start_step_function(DELETE_ROLE_SFN_ARN, account_number, event_name, role_name)

        else:
            logger.info(f"Ignoring unsupported eventName: {event_name}")
            return {"status": "ignored", "eventName": event_name}

    except Exception as e:
        logger.error(f"Unhandled exception: {e}", exc_info=True)
        raise

def start_step_function(state_machine_arn, account_number, event_name, role_name):
    """
    Starts an AWS Step Function execution.
    """
    if not state_machine_arn:
        logger.error(f"Missing Step Function ARN for event: {event_name}")
        raise ValueError(f"Missing Step Function ARN for event: {event_name}")

    input_payload = {
        "account": account_number,
        "eventName": event_name,
        "roleName": role_name
    }

    try:
        response = sfn_client.start_execution(
            stateMachineArn=state_machine_arn,
            input=json.dumps(input_payload)
        )
        logger.info(f"Step Function '{event_name}' started: {response['executionArn']}")
        return {"status": "started", "executionArn": response['executionArn']}

    except ClientError as e:
        logger.error(f"Failed to start Step Function for event {event_name}: {e}")
        raise