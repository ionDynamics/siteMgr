{{template "header" .}}
{{$cinf := .Name | clientInfo}}
{{with .Name}}
		<div class="row">
			<div class="col-xs-12 col-sm-6" style="text-align: right;"><br>
				<p>
					Currently logged in as <b>{{.}}</b><br>
					Your Masterpassword generated this image:<br>
					If it looks different than usual you misspelled it maybe
				</p>
			</div>
			<div class="col-xs-12 col-sm-2">
				{{$cinf.Identicon}}
			</div>
			<div class="col-xs-12 col-sm-4">
				<p class="small text-muted">
					<b>Client Info</b><br>
					Message Version: {{$cinf.MsgVersion}}<br>
					<table class="clientInfo">
						<tr>
							<td>Vendor:&nbsp;</td><td>{{$cinf.Vendor}}</td>
						</tr>
						<tr>
							<td>Client:&nbsp;</td><td>{{$cinf.Client}}</td>
						</tr>
						<tr>
							<td>Variant:&nbsp;</td><td>{{$cinf.Variant}}</td>
						</tr>
						<tr>
							<td>Address:&nbsp;</td><td>{{$cinf.Address}}</td>
						</tr>
					</table>
				</p>
			</div>
		</div>				
{{end}}



<div class="row">
	<div class="col-xs-12">

		<h2>My Sites</h2>

		<table class="table table-bordered table-hover table-striped">
			<tr>
				<th>Name</th>
				<th>Version</th>
				<th>Template</th>
				<th>Login</th>
				<th>Email</th>
				<th>&nbsp;</th>
			</tr>
			{{range .GetSites}}
			<tr>
				<td>
					{{.Name}}
				</td>
				<td>
					{{.Version}}
				</td>
				<td>
					{{ if eq .Template "pppppppp"}} Alphanumeric8{{else}}
					{{ if eq .Template "pppppppppp"}} Alphanumeric10{{else}}
					{{ if eq .Template "pppppppppppppppp"}} Alphanumeric16{{else}}
					{{ if eq .Template "pppppppppppppppppppp"}} Alphanumeric20{{else}}
					{{ if eq .Template "xxxxxxxxxxxxxxxx"}} Printable16{{else}}
					{{ if eq .Template "xxxxxxxxxxxxxxxxxxxxxxxxx"}} Printable25{{else}}
					{{ if eq .Template "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}} Printable32{{else}}{{.Template}} {{end}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}
				</td>
				<td>
					{{with .Login}}						
						{{.}}
						<form action="/clip/send" method="post" class="pull-right"><input type="hidden" name="clip-content" value="{{.}}"><button class="btn btn-default" type="submit"><span class="glyphicon glyphicon-copy"></span></button></form>
					{{end}}
				</td>
				<td>
					{{with .Email}}
						{{.}}
						<form action="/clip/send" method="post" class="pull-right"><input type="hidden" name="clip-content" value="{{.}}"><button class="btn btn-default" type="submit"><span class="glyphicon glyphicon-copy"></span></button></form>
					{{end}}
				</td>
				<td>
					<form action="/site/send" method="post">
						<input type="hidden" name="site-name" value="{{.Name}}">
						<button class="btn btn-primary pull-left" type="submit">Password</button>
					</form>

					<form action="/site/del" method="post">
						<input type="hidden" name="site-name" value="{{.Name}}">
						<button class="btn pull-left delBtn" x-data-del="{{.Name}}" type="submit">Delete</button>
						<button class="btn pull-right btn-danger hidden confirmBtn" type="submit">Confirm</button>
					</form>
				</td>
			</tr>
			{{end}}
		</table>
	</div>
</div>
{{if atLeast "0.6.0" $cinf.MsgVersion}}
<div class="row">
	<div class="col-xs-12">

		<h2>My Credentials</h2>

		<table class="table table-bordered table-hover table-striped">
			<tr>
				<th>Name</th>
				<th>Login</th>
				<th>Email</th>
				<th>&nbsp;</th>
			</tr>
			{{range .GetAllCredentials}}
			<tr>
				<td>
					{{.Name}}
				</td>
				<td>
					{{with .Login}}						
						{{.}}
						<form action="/clip/send" method="post" class="pull-right"><input type="hidden" name="clip-content" value="{{.}}"><button class="btn btn-default" type="submit"><span class="glyphicon glyphicon-copy"></span></button></form>
					{{end}}
				</td>
				<td>
					{{with .Email}}
						{{.}}
						<form action="/clip/send" method="post" class="pull-right"><input type="hidden" name="clip-content" value="{{.}}"><button class="btn btn-default" type="submit"><span class="glyphicon glyphicon-copy"></span></button></form>
					{{end}}
				</td>
				<td>
					<form action="/credentials/send" method="post">
						<input type="hidden" name="credentials-name" value="{{.Name}}">
						<button class="btn btn-primary pull-left" type="submit">Password</button>
					</form>

					<form action="/credentials/del" method="post">
						<input type="hidden" name="credentials-name" value="{{.Name}}">
						<button class="btn pull-left delBtn" x-data-del="{{.Name}}" type="submit">Delete</button>
						<button class="btn pull-right btn-danger hidden confirmBtn" type="submit">Confirm</button>
					</form>
				</td>
			</tr>
			{{end}}
		</table>
	</div>
</div>
{{end}}

<div class="row">
	<div class="col-xs-12 col-sm-6 col-sm-offset-3 col-md-offset-0">
		<div class="well">
			<h3>New Site:</h3>
			<form action="/site/set" method="post">
				<div class="form-group">
					<input type="text" class="form-control" name="site-name" placeholder="Site URL" required>
				</div>				
				<div class="form-group">
					<input type="text" class="form-control" name="site-login" placeholder="Login">
				</div>
				<div class="form-group">
					<input type="text" class="form-control" name="site-email" placeholder="E-Mail">
				</div>
				<div class="form-group">
					<input type="text" class="form-control" name="site-version" placeholder="Password Version" value="1">
				</div>
				<div class="form-group">
					<select name="site-template" class="form-control">
					<option value="pppppppp">Alphanumeric8</option>
					<option value="pppppppppp">Alphanumeric10</option>
					<option value="pppppppppppppppp" selected="selected">Alphanumeric16</option>
					<option value="pppppppppppppppppppp">Alphanumeric20</option>
					<option value="xxxxxxxxxxxxxxxx">Printable16</option>
					<option value="xxxxxxxxxxxxxxxxxxxxxxxxx">Printable25</option>
					<option value="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx">Printable32</option>
					</select>
					<!--<input type="text" class="form-control" name="site-template" placeholder="Password Template">-->
				</div>
				<div class="form-group">
					<button class="btn btn-primary" type="submit">Add Site</button>
				</div>
			</form>
		</div>
	</div>
	{{if atLeast "0.6.0" $cinf.MsgVersion}}
		<div class="col-xs-12 col-sm-6 col-sm-offset-3 col-md-offset-0">
			<div class="well">
				<h3>New Credentials:</h3>
				<form action="/credentials/set" method="post" autocomplete="off">
					<div class="form-group">
						<input type="text" class="form-control" name="credentials-name" placeholder="Site URL" required>
					</div>		
					<div class="form-group">
						<input type="text" class="form-control" name="credentials-login" placeholder="Login">
					</div>
					<div class="form-group">
						<input type="text" class="form-control" name="credentials-email" placeholder="E-Mail" autocomplete="off">
					</div>
					<div class="form-group">
						<input type="password" class="form-control" name="credentials-pass" placeholder="Password" autocomplete="off">
					</div>
					<div class="form-group">
						<input type="password" class="form-control" name="credentials-pass-repeat" placeholder="Repeat Password" autocomplete="off">
					</div>
					<div class="form-group">
						<button class="btn btn-primary" type="submit">Add Credentials</button>
					</div>
				</form>
			</div>
		</div>
	{{end}}
</div>

{{template "footer" .}}