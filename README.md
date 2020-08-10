# Mass
Small utility to stress-test your DNS infrastructure. Working only on Linux-based systems.

⚠️**The author is not responsible for any harm caused by third parties using this code**⚠️

## Install
Install from binary:
```shell
wget -c https://github.com/velp/mass/releases/download/v0.0.1rc1/mass_0.0.1rc1_Linux_x86_64.tar.gz -O - | tar -xz
```

Install from source

```shell
go get github.com/velp/mass
cd $GOPATH/src/github.com/velp/mass/
make build
```

## Run tests
### DNS A query flood + src IP spoofing
How to start DNS-flood test with source IP spoofing (target host `192.168.0.3`, target domain `kokoko.ru`):

```shell
mass -dst-ip=192.168.0.3 -dns-domain="kokoko.ru" -src-ip-range="172.16.10-40.1-255"
```

this command will start DNS A query flooding from IP addresses 172.16.10.1-172.16.40.255.

Incoming traffic will look like:

```
14:38:09.504415 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 69: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 55)
    172.16.32.156.58346 > 192.168.0.3.53: [udp sum ok] 43690+ A? kokoko.ru. (27)
14:38:09.504416 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 69: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 55)
    172.16.32.154.41621 > 192.168.0.3.53: [udp sum ok] 43690+ A? kokoko.ru. (27)
14:38:09.504417 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 69: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 55)
    172.16.32.157.61968 > 192.168.0.3.53: [udp sum ok] 43690+ A? kokoko.ru. (27)
14:38:09.504418 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 69: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 55)
    172.16.32.158.58003 > 192.168.0.3.53: [udp sum ok] 43690+ A? kokoko.ru. (27)
```

### DNS A query randomization
How to start test with random part of the domain (target host `192.168.0.3`, target domain `kokoko.ru`):

```shell
mass -dst-ip=192.168.0.3 -dns-domain="*.kokoko.ru"
```

Incoming traffic will look like:

```
14:41:04.740469 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 80: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 66)
    192.168.0.2.50580 > 192.168.0.3.53: [udp sum ok] 43690+ A? arYDwKdbxO.kokoko.ru. (38)
14:41:04.743010 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 80: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 66)
    192.168.0.2.30935 > 192.168.0.3.53: [udp sum ok] 43690+ A? BmNiYvHeDR.kokoko.ru. (38)
14:41:04.743038 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 80: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 66)
    192.168.0.2.31416 > 192.168.0.3.53: [udp sum ok] 43690+ A? KTuMYrsVht.kokoko.ru. (38)
14:41:04.743078 fa:16:3e:81:70:04 > fa:16:3e:08:fc:d3, ethertype IPv4 (0x0800), length 80: (tos 0x0, ttl 64, id 0, offset 0, flags [none], proto UDP (17), length 66)
    192.168.0.2.60845 > 192.168.0.3.53: [udp sum ok] 43690+ A? yfBXxhEngL.kokoko.ru. (38)
```

### DNS A queries to check DNS health
To run `mass` in checker mode you have to add argument `-module=dns-checker`:

```shell
mass -dst-ip=8.8.8.8 -dns-domain="selectel.ru" -module=dns-checker
```

The utility will use real public IP address and generate real A DNS queries. As result you will see small report in stdout:

```shell
mass -dst-ip=8.8.8.8 -dns-domain="selectel.ru" -module=dns-checker
2020/08/10 14:47:25 use gateway IP 192.168.0.1 to fnd destination MAC address because IP address 8.8.8.8 is out of broadcast domain 192.168.0.2/24
2020/08/10 14:47:26 Send ARP request who-has 192.168.0.1 tell 192.168.0.2
2020/08/10 14:47:26 IP 192.168.0.1 is at fa:16:3e:94:fe:4a
2020/08/10 14:47:26 Network data:
	Source: 192.168.0.2 (fa:16:3e:81:70:04, real IP 192.168.0.2) from eth0
	Destination: 8.8.8.8 (fa:16:3e:94:fe:4a)
2020/08/10 14:47:26 DNS checker #0 started
2020/08/10 14:47:26 DNS checker #2 started
2020/08/10 14:47:26 DNS checker #1 started
2020/08/10 14:47:26 DNS checker #8 started
2020/08/10 14:47:26 DNS checker #6 started
2020/08/10 14:47:26 DNS checker #7 started
2020/08/10 14:47:26 DNS checker #3 started
2020/08/10 14:47:26 DNS checker #4 started
2020/08/10 14:47:26 DNS checker #9 started
2020/08/10 14:47:26 DNS checker #5 started
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 15.831688ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 17.690128ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 17.639631ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 17.722368ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 18.274828ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 20.696157ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 21.053188ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 20.942986ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 30.956754ms)
2020/08/10 14:47:27 ip addresses: 95.213.255.1  (request took: 30.608542ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 15.040929ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 15.317122ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 15.270774ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 16.758337ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 17.096351ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 17.13202ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 17.130884ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 16.950046ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 21.683803ms)
2020/08/10 14:47:28 ip addresses: 95.213.255.1  (request took: 31.068792ms)
^C2020/08/10 14:47:28 Signal received attempting
2020/08/10 14:47:28 DNS checker #3 stoped
2020/08/10 14:47:28 DNS checker #9 stoped
2020/08/10 14:47:28 DNS checker #7 stoped
2020/08/10 14:47:28 DNS checker #0 stoped
2020/08/10 14:47:28 DNS checker #4 stoped
2020/08/10 14:47:28 DNS checker #2 stoped
2020/08/10 14:47:28 DNS checker #1 stoped
2020/08/10 14:47:28 DNS checker #6 stoped
2020/08/10 14:47:28 DNS checker #8 stoped
2020/08/10 14:47:28 DNS checker #5 stoped
2020/08/10 14:47:28 Total requests: 20 failed: 0 success: 20
```

## Performance
DNS random queries + IP spoofing, 10 goroutines:

```shell
mass -dst-ip=192.168.0.3 -dns-domain="*.kokoko.ru" -src-ip-range="172.16.10-40.1-255
```

Result for 5000000 packets:

```shell
File name:           ./dns.pcap
File type:           Wireshark/tcpdump/... - pcap
File encapsulation:  Ethernet
File timestamp precision:  microseconds (6)
Packet size limit:   file hdr: 262144 bytes
Number of packets:   5000 k
File size:           480 MB
Data size:           400 MB
Capture duration:    26.256245 seconds
First packet time:   2020-08-10 15:03:54.140370
Last packet time:    2020-08-10 15:04:20.396615
Data byte rate:      15 MBps
Data bit rate:       121 Mbps
Average packet size: 80.00 bytes
Average packet rate: 190 kpackets/s
SHA256:              1beb9aa357ebc530ae9a834c268dc164b96a79cc23d791f70583343a54fe0e05
RIPEMD160:           ad7492aee1613ba4ba30d247f812e897ae58a329
SHA1:                94cb4dc595e6ff9a7052f5fb284a588436a40716
Strict time order:   True
Number of interfaces in file: 1
Interface #0 info:
                     Encapsulation = Ethernet (1 - ether)
                     Capture length = 262144
                     Time precision = microseconds (6)
                     Time ticks per second = 1000000
                     Number of stat entries = 0
                     Number of packets = 5000000
```
