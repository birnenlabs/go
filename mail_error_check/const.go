package main

const temporaryErrorTmpl = `Mail Delivery Failure

This is an automatically generated Delivery Status Notification

THIS IS A WARNING MESSAGE ONLY.
YOU DO NOT NEED TO RESEND YOUR MESSAGE.

Delivery to the following recipient has been delayed:

     %s

Technical details of temporary failure:
%s
`

const permanentErrorTmpl = `Mail Delivery Permanent Failure

This is an automatically generated Delivery Status Notification

Delivery to the following recipient failed permanently:

     %s

Technical details of permanent failure:

%s
`

const technicalDetailsTmpl = `========== Attempt %d ==========
Timestamp: %s
Severity: %s
SMTP error code: %d
Reason: %s
%s

`
