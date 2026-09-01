#!/bin/tclsh
load tclrega.so

proc url_decode {str} {
  set str [string map -nocase [list "+" " "] $str]
  set result ""
  while {[regexp -indices {%([0-9A-Fa-f]{2})} $str match]} {
    lassign $match start end
    append result [string range $str 0 [expr {$start - 1}]]
    set hex [string range $str [expr {$start + 1}] $end]
    append result [format "%c" 0x$hex]
    set str [string range $str [expr {$end + 1}] end]
  }
  append result $str
  return $result
}

proc get_sid {} {
  global env
  set sid ""
  catch {
    set input $env(QUERY_STRING)
    set pairs [split $input &]
    foreach pair $pairs {
      if {0 != [regexp "^sid=(@.*@)$" $pair dummy val]} {
        set sid $val
        break
      }
    }
  }
  return $sid
}

proc check_session sid {
  if {[regexp {@([0-9a-zA-Z]{10})@} $sid all sidnr]} {
    set res [lindex [rega_script "Write(system.GetSessionVarStr('$sidnr'));"] 1]
    if {$res != ""} {
      return 1
    }
  }
  return 0
}

proc read_config {config_file varname_apikey varname_transport varname_port} {
  upvar 1 $varname_apikey apikey
  upvar 1 $varname_transport transport
  upvar 1 $varname_port port
  set apikey ""
  set transport ""
  set port ""
  if {[file exists $config_file]} {
    set fh [open $config_file r]
    set content [read $fh]
    close $fh
    regexp {apiKey\s*=\s*'([^']*)'} $content dummy apikey
    regexp {transport\s*=\s*'([^']*)'} $content dummy transport
    regexp {port\s*=\s*([0-9]+)} $content dummy port
  }
}

proc save_config {config_file new_apikey new_transport} {
  if {![file exists $config_file]} {
    return 0
  }
  set fh [open $config_file r]
  set content [read $fh]
  close $fh
  if {$new_apikey ne ""} {
    regsub {apiKey\s*=\s*'[^']*'} $content "apiKey = '$new_apikey'" content
  }
  if {$new_transport ne ""} {
    regsub {transport\s*=\s*'[^']*'} $content "transport = '$new_transport'" content
  }
  set fh [open $config_file w]
  puts -nonewline $fh $content
  close $fh
  return 1
}

proc handle_post {config_file varname_apikey varname_transport} {
  global env saved_msg
  upvar 1 $varname_apikey apikey
  upvar 1 $varname_transport transport
  set saved_msg 0
  if {![info exists env(REQUEST_METHOD)] || $env(REQUEST_METHOD) ne "POST"} {
    return
  }
  set cl $env(CONTENT_LENGTH)
  set post_data [read stdin $cl]
  set new_apikey ""
  set new_transport ""
  foreach pair [split $post_data &] {
    if {[regexp {^apikey=(.*)$} $pair dummy val]} {
      set new_apikey [url_decode $val]
    }
    if {[regexp {^transport=(.*)$} $pair dummy val]} {
      set new_transport [url_decode $val]
    }
  }
  if {$new_apikey eq "" && $new_transport eq ""} {
    return
  }
  if {[save_config $config_file $new_apikey $new_transport]} {
    if {$new_apikey ne ""} { set apikey $new_apikey }
    if {$new_transport ne ""} { set transport $new_transport }
    set saved_msg 1
  }
}

proc print_form {sid apikey transport port} {
  global saved_msg
  puts "<form method=\"post\" action=\"config.cgi?sid=$sid\">"
  puts "<div class=\"form-group\">"
  puts "<label for=\"apikey\">API-Key:</label>"
  puts "<input type=\"text\" id=\"apikey\" name=\"apikey\" value=\"$apikey\">"
  puts "</div>"
  puts "<div class=\"form-group\">"
  puts "<label for=\"transport\">Transport:</label>"
  set sel_http ""
  set sel_https ""
  if {$transport eq "http"} { set sel_http " selected" }
  if {$transport eq "https"} { set sel_https " selected" }
  puts "<select id=\"transport\" name=\"transport\"><option value=\"http\"$sel_http>HTTP</option><option value=\"https\"$sel_https>HTTPS</option></select>"
  puts "</div>"
  puts "<input type=\"submit\" value=\"Änderungen speichern\" class=\"btn\">"
  puts "</form>"
  if {$saved_msg} {
    puts "<p id=\"msg\">Die Konfiguration wurde gespeichert.</p>"
    puts "<script>setTimeout(function(){var e=document.getElementById('msg');if(e)e.style.display='none';},3000);</script>"
  }
  puts "<h2>MCP-Client / Agent Konfiguration</h2>"
  puts "<table>"
  puts "<tr><th>URL</th><td>${transport}://&lt;ccu-ip&gt;:${port}/mcp</td></tr>"
  puts "<tr><th>Transport</th><td>${transport}</td></tr>"
  puts "<tr><th>Port</th><td>${port}</td></tr>"
  puts "<tr><th>API-Key</th><td>${apikey}</td></tr>"
  puts "<tr><th>Authentifizierung</th><td>Bearer Token (API-Key)</td></tr>"
  puts "</table>"
}

proc print_html_header {} {
  puts "Content-Type: text/html; charset=utf-8"
  puts ""
  puts "<!doctype html><html><head><meta charset=\"utf-8\"><title>CCU-AI-MCP Konfiguration</title>"
  puts "<style>"
  puts "body { font-family: sans-serif; background: #e0f7f7; color: #333; margin: 0; padding: 40px 20px; }"
  puts "h1 { color: #008080; border-bottom: 3px solid #008080; padding-bottom: 10px; }"
  puts ".container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 30px; box-shadow: 0 2px 8px rgba(0,128,128,0.15); }"
  puts "form { margin-top: 20px; }"
  puts ".form-group { margin-bottom: 20px; }"
  puts "label { display: block; font-weight: bold; margin-bottom: 6px; color: #006666; }"
  puts "input\[type=text\], select { width: 100%; padding: 10px; border: 1px solid #b0d8d8; border-radius: 4px; font-size: 14px; box-sizing: border-box; }"
  puts "input\[type=text\]:focus, select:focus { border-color: #008080; outline: none; box-shadow: 0 0 4px rgba(0,128,128,0.3); }"
  puts ".btn { background: #008080; color: #fff; border: none; padding: 10px 24px; border-radius: 4px; font-size: 14px; cursor: pointer; }"
  puts ".btn:hover { background: #006666; }"
  puts "p { margin-top: 15px; padding: 10px; background: #d4f0f0; border-radius: 4px; color: #006666; }"
  puts "h2 { color: #008080; margin-top: 30px; border-bottom: 2px solid #b0d8d8; padding-bottom: 8px; }"
  puts "table { width: 100%; border-collapse: collapse; margin-top: 15px; }"
  puts "th { text-align: left; padding: 8px 12px; background: #e0f7f7; color: #006666; border: 1px solid #b0d8d8; width: 40%; }"
  puts "td { padding: 8px 12px; border: 1px solid #b0d8d8; font-family: monospace; }"
  puts "</style>"
  puts "</head><body><div class=\"container\"><h1>CCU-AI-MCP Konfiguration</h1>"
}

proc print_html_footer {} {
  puts "</div></body></html>"
}

proc print_not_authenticated {} {
  puts "<p>Der Zugriff auf die Konfigurationsoberfläche ist nur über eine Anmeldung an der Web-UI der CCU möglich!</p>"
}

set sid [get_sid]
set config_file "/usr/local/addons/ccu-ai-mcp/config.toml"

print_html_header

if {[info exists sid] && [check_session $sid]} {
  read_config $config_file apikey transport port
  handle_post $config_file apikey transport
  print_form $sid $apikey $transport $port
} else {
  print_not_authenticated
}

print_html_footer
