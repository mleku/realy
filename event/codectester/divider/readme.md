# divider

this is a subset of the codectester, its purpose is to accept a large volume of (ostensibly valid) event data in a 
`.jsonl` file and categorize it into subsets that have failed at separate parts of the processing.

it generates 6 `.jsonl` output files:

- `read.jsonl` - simply makes a copy of the input
- `fail_unmar.jsonl` - writes all of the lines of the input that fail to unmarshal
- `fail_ids.jsonl` - the events are re-marshaled into their canonical form, and if the hash of the canonical form of the 
   event does not match the one in the object form of the event, the event (from its original form) is written as a line 
   in this.
- `fail_tobin.jsonl` - events that failed to be rendered into the binary form.
- `fail_frombin.jsonl` - events that failed to be returned to the runtime form.
- `fail_reser.jsonl` - events that failed to retain all of the data successfully decoded from the original JSON form based on
   the difference between the way the JSON encoding from the runtime version is different after all of the above steps
   have been run.

## purpose

this is as a way to isolate the categories of encoding/decoding errors in order to make the event codec as robust and 
correct as possible.