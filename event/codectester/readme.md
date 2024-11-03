# codectester

a simple tool that takes a file, checks that both the JSON and binary codecs are correctly 
encoding and successfully recover an event going between the source JSON, to the runtime form,
to the binary form, then back to the runtime form, and back to the JSON canonical form, and 
produces the same canonical event ID hash as the original event.