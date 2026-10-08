package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	s := `
<div class="container">
        <div class="row">
            <div class="col-xs-12 col-md-9 main">
                <div class="visible-lg">
                    <div class="side-widget">
                        <i class="fa fa-bookmark item " id="side-widget-bookmarks-btn"></i>
                        <i class="fa fa-weibo item"></i>
                        <i class="fa fa-weixin item" data-toggle="popover" data-placement="right"></i>
                        <i class="fa fa-twitter item"></i>
                        <i class="fa fa-facebook item"></i>
                        <i class="fa fa-arrow-up item hidden"></i>
                    </div>
                </div>
                
![]()

>>>>>>>>>>>>>>>>>
`

	type a struct {
		S string
	}

	b := a{s}
	raw, err := json.Marshal(b)
	fmt.Printf("%#v,%#v", string(raw), err)
}
