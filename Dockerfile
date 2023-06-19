FROM alpine
ADD match_process /match_process
ENTRYPOINT [ "/match_process" ]
