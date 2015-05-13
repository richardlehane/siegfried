import sys
import fpr_pb2

response = early_adopter_create_Fpr_stub('localhost', 5138).Identify(fpr_pb2.Request(path=sys.argv[1]), 10)
if reponse.error == '':
    print >> sys.stderr, response.error
    exit(1)
print response.puid

