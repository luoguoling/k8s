FROM centos:8
MAINTAINER ROLIN
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
#RUN yum install -y gcc gcc-c++ make gd-devel libxml2-devel libcurl-devel libjpeg-devel libpng-devel openssl-devel
RUN yum -y install gcc gcc-c++ openssl openssl-devel zlib zlib-devel make
#安装yum源
RUN rpm -Uvh https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm 
RUN rpm -Uvh https://mirror.webtatic.com/yum/el7/webtatic-release.rpm
#安装nginx
RUN yum -y install nginx
COPY nginx.conf /etc/nginx.conf
COPY www.conf /etc/nginx/conf.d/
RUN mkdir -p /data/website
#安装php-fpm
RUN yum install -y php php-devel php-fpm php-cli
RUN php -m
RUN yum install -y php-pecl-igbinary --skip-broken
RUN yum install -y php-bz2 php-soap  php-zip php-opcache php-xmlrpc php-calendar  php-json php-bcmath php-exif php-dom php-gettext \
php-gd php-json php-mbstring php-mysqli php-mysqlnd php-pdo php-pdo_mysql php-pdo_sqlite php-readline  php-posix php-redis php-simplexml  \
php-soap php-sqlite3  php-xml php-xmlreader php-xmlrpc php-xmlwriter php-zip
RUN sed -i -e 's@;pid = run/php-fpm.pid@pid = run/php-fpm.pid@g' -e 's@nobody@php@g' -e 's@listen = 127.0.0.1:9000@listen = 0.0.0.0:9000@g' /etc/php-fpm.conf
RUN sed -i 's@;daemonize = yes@daemonize = no@g' /etc/php-fpm.conf
RUN  rm /etc/php-fpm.d/www.conf
COPY fpm.conf /etc/php-fpm.d/
COPY php.ini /etc/php.ini
COPY php-fpm.conf /etc/php-fpm.conf
RUN pecl install swoole
#安装compose
RUN php -r "copy('https://install.phpcomposer.com/installer', 'composer-setup.php');"
RUN php composer-setup.php
RUN php -r "unlink('composer-setup.php');"
RUN mv composer.phar /usr/local/bin/composer
RUN mkdir /run/php-fpm -p
RUN yum clean all
RUN yum -y remove gcc gcc-c++ make
EXPOSE 9000
#CMD ["php-fpm","-c","/etc/php-fpm.conf"]
#CMD ["php-fpm"]
#CMD ["/usr/sbin/nginx","-g","daemon off;"]
CMD php-fpm && nginx -g "daemon off;"
