import ctypes
so = ctypes.cdll.LoadLibrary('./_checksig.so')
verify = so.verify
verify.argtypes = [ctypes.c_char_p]
verify.restype = ctypes.c_void_p
free = so.free
free.argtypes = [ctypes.c_void_p]
ptr = verify('/tmp/logs'.encode('utf-8'))
out = ctypes.string_at(ptr)
free(ptr)
print(out.decode('utf-8'))
# "/tmp/logs/httpd-08.log" - mismatch
