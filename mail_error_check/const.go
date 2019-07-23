package main

const temporaryErrorTmpl = `Mail Delivery %s Failure

This is an automatically generated Delivery Status Notification

THIS IS A WARNING MESSAGE ONLY.
YOU DO NOT NEED TO RESEND YOUR MESSAGE.

Delivery to the following recipient:

     %s

has been delayed.

This was %d attempt, message will be retried up to 8 times.

Technical details of temporary failure:
SMTP error code: %d
Reason: %s
%s
`

const permanentErrorTmpl = `Mail Delivery %s Failure

This is an automatically generated Delivery Status Notification

Delivery to the following recipient:

     %s

failed permanently after %d attempt(s).

Technical details of permanent failure:
SMTP error code: %d
Reason: %s
%s
`
