    # dmonitord

dmonitord is a daemon for monitor domains expire date.


# How to run

    go get github.com/domainr/whois
    go get github.com/forease/ebase
    git clone https://github.com/jonsen/dmonitord
    cd dmonitord
    ./mk.sh
    ./run.sh

# Setup

## dmonitord.conf
    vi etc/dmonitord.conf
    
    [common]
    adminer = "mymail@mail.com,youmail@mail.com"
    dfile = "./etc/domains.txt"
    retry = 3
    
    [log]
    enable = true
    # Critical = 0, Error = 1, Warning = 2, Info = 3, Debug = 4, Trace = 5
    level = 5
    # syslog, consloe, file
    type = "consloe"
    file = ""
    
    [smtp]
    host = "smtp.exmail.com"
    port = 25
    user = "sender@sendmailer.com"
    password = "mailpassword"
    auth = true
    tls  = false


## domains.txt

    vi domains.txt
    
    google.com
    facebook.com
    ibm.com


# Contact me

    Jonsen Yang
    im16hot#gmail.com (replace # with @)
