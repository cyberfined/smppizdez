# smppizdez

Just another fucking SMPP client with following features supported:

1) GSM7 (packed/unpacked) and UCS2 encodings.
2) Message Payload TLV.
3) UDH segmentation.
4) SAR segmentation.
5) Non-standard segment sizes (any between 1 and 255).
6) Deceptive encoding (you encode message with for example GSM7 but send UCS2 in data_coding field instead).
7) TLV.

# TODO

1) Add TLS support.
2) Fix message field indication of incorrectness (for some reason border color is not applied to Gtk.TextView).
3) Fix account dialog (it must become focused when opened).
4) Support DATA_SM.
5) Make more beautiful PDU logs.
6) Use more convenient path to store json file with accounts, like ~/.config/smppizdez/data.json
