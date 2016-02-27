{{template "header" .}}
	<div class="row">
		<div class="col-sm-12 col-md-5 col-lg-4">
			<div class="well">
				<h3><span class="glyphicon glyphicon-import"></span>&nbsp;&nbsp;&nbsp;Backup</h3><br>
				<div class="row">
					<div class="form-group col-xs-12 col-md-9">
						<a href="/backup/get"><button class="btn btn-default">Download</button></a>
					</div>
				</div>

				<div><h4><span class="label label-default">This will export your sites to your computer</span></h4></div>
				
			</div>
		</div>
		<div class="col-sm-12 col-md-7 col-lg-8">
			<div class="well">
				<h3><span class="glyphicon glyphicon-export"></span>&nbsp;&nbsp;&nbsp;Recover</h3><br>
				<form action="/backup/recover" method="post" enctype="multipart/form-data">					
					<div class="row" >
						<div class="form-group col-xs-12 col-md-10">
							<input type="file" class="form-control" name="recover-json" placeholder="Upload File" required>
						</div>
						<div class="form-group col-xs-12 col-md-2">
							<button class="btn btn-default pull-right" type="submit">Upload</button>
						</div>
					</div>
					<div>
						<h4><span class="label label-default">This will overwrite identical named sites</span></h4>
					</div>
				</form>
			</div>
		</div>
	</div>
{{template "footer" .}}