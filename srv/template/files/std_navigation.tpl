{{define "navigation"}}
		<nav class="navbar navbar-default navbar-fixed-top">
			<div class="container">
				<div class="navbar-header">
					<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar" aria-expanded="false" aria-controls="navbar">
						<span class="sr-only">Toggle navigation</span>
						<span class="icon-bar"></span>
						<span class="icon-bar"></span>
						<span class="icon-bar"></span>
					</button>
					<a class="navbar-brand" href="/">SiteMgr</a>
				</div>
				<div id="navbar" class="collapse navbar-collapse">
					<ul class="nav navbar-nav navbar-right">
					{{with .}}
						<li class="pull-right"><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Log out</a></li>
						<li class="pull-right"><a href="/backup/mgr"><span class="glyphicon glyphicon-hdd"></span> Backup</a></li>
					{{else}}
						<li class="pull-right"><a href="/register"><span class="glyphicon glyphicon-log-in"></span> Register</a></li>
					{{end}}
					</ul>
				</div><!--/.nav-collapse -->
			</div>
	    </nav>
{{end}}