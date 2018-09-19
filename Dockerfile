FROM centos

ADD entrypoint.sh /entrypoint.sh
add vnet-bgp-monitor /vnet-bgp-monitor
ENTRYPOINT ["/entrypoint.sh"]


