# Pull base image.
# NOTE: to use this Dockerfile, mount the chain data, including an init.sh, to /home/eris/data 
FROM quay.io/eris/base


ENV TMROOT /home/$USER/data
WORKDIR $TMROOT
RUN chown -R $USER:$USER $TMROOT

# ports: 1337:eris-db API; 46656:mint-peer; 46657:mint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657

# Run tendermint
CMD ["./init.sh"] # 
