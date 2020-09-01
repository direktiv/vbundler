#!/vorteil/busybox ash 

binary=/vorteil/busybox

list=`$binary --list`

busybox_links()
{
        dir=$1
        $binary mkdir -p $dir
        for command in $list; do 
                $binary ln -s -f $binary $dir/$command
        done
}

busybox_links /bin
busybox_links /usr/bin
