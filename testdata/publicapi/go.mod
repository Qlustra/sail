module sail-quickstart

go 1.27

require github.com/qlustra/sail v1.0.0

require github.com/qlustra/intel v1.0.0 // indirect

// Exercise the Sail checkout before it has a tag. Not needed in consumer modules.
replace github.com/qlustra/sail => ../..
