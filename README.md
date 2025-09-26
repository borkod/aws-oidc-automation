# aws-oidc-automation

This repository accompanies the presentation at the Commit Your Code conference. The presentation can be found [here](https://github.com/borkod/CYC2025).

It includes Terraform and Lambda source code to set up all resources required for the OIDC automation account configuration.

## The Problem

Imagine you have an external application (let's say a CI/CD system) that needs to interact with your AWS account. This could be for deploying applications, updating resources, or provisioning new infrastructure. No big deal, right?

But here's the kicker: this CI/CD system should run on autopilot, with minimal user input. It might kick off pipelines in the early hours based on a schedule, trigger a deployment when a PR is merged, or even run in the middle of the night to respond to an incident.

So, how do you ensure the system has the right permissions to carry out these tasks?

Enter IAM Users... the Old-School Way

A quick and straightforward solution would be to create an IAM user for the CI/CD system, generate an Access Key ID and Secret Access Key, and let it run free. That’s one way to get the job done.

But here’s the catch: Access Key ID and Secret Access Key are essentially a static username/password combo. They’re long-lived credentials—meaning they don’t expire unless you manually rotate them—and they’re a major security risk if exposed.

In other words, if these credentials are ever leaked, your AWS environment is wide open to exploitation. And that’s not something you want, especially for a system running unattended.

The Quick Fix: Rotate Credentials... But Is It Enough?

Many organizations go down the road of building Frankenstein-like scripts to rotate Access Key ID/Secret Access Keys regularly. While this can help mitigate the risk of exposed credentials, it’s more of a band-aid solution than a sustainable fix.

So, what's the better way to handle this situation, you ask?

Let’s dive into a more secure, scalable approach!