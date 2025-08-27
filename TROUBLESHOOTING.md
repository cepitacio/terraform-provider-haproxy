# Troubleshooting

This document provides solutions to common issues you may encounter while using the HAProxy Terraform provider.

## Common Issues

### "Transaction outdated" or "Version mismatch" errors

This error occurs when the HAProxy configuration has been changed by another process between the time the provider reads the configuration and when it tries to apply the changes.

**Solution:**

-   Ensure that no other processes are modifying the HAProxy configuration while you are running Terraform.
-   If you are running multiple instances of Terraform at the same time, ensure that they are not modifying the same HAProxy instance.

### "Permission denied" errors

This error occurs when the user running Terraform does not have the necessary permissions to access the HAProxy Data Plane API.

**Solution:**

-   Ensure that the user running Terraform has the necessary permissions to access the HAProxy Data Plane API.
-   If you are using a username and password to authenticate with the HAProxy Data Plane API, ensure that they are correct.

## Debugging

If you are still having issues, you can enable debug logging to get more information about what is happening. To enable debug logging, set the `TF_LOG` environment variable to `DEBUG`.

```shell
export TF_LOG=DEBUG
```

## Reporting Issues

If you are still unable to resolve the issue, please open an issue on our GitHub repository. Please include the following information:

-   A clear and descriptive title.
-   A detailed description of the bug, including steps to reproduce it.
-   The version of the provider you are using.
-   Any relevant configuration files or code snippets.
-   The output of the `terraform apply` command with debug logging enabled.