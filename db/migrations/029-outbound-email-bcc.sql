-- Record the BCC audit copy on every outbound send.
ALTER TABLE outbound_email ADD COLUMN bcc_addrs TEXT NOT NULL DEFAULT '';
