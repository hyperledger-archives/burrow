import sys
import re

masterfile = sys.argv[1]
print (masterfile)

outfile = None;

beginRE = re.compile(r"^\s*\/\/\s*#\s*begin\s*\(\s*(\w+)\s*\)", re.I)

for line in open(masterfile, 'r'):
	match = beginRE.search(line)
	if match is not None and match.group(1) is not "":
		outfile = open(match.group(1) + ".proto", 'w')
		outfile.write("syntax = 'proto3';\n\nimport \"common.proto\";\n\n")
	elif outfile is not None:
		outfile.write(line)
	else:
		print "skip"

print "done"

