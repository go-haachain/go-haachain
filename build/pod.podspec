Pod::Spec.new do |spec|
  spec.name         = 'Ghaa'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/haachain/go-haachain'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS haachain Client'
  spec.source       = { :git => 'https://github.com/haachain/go-haachain.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Ghaa.framework'

	spec.prepare_command = <<-CMD
    curl https://ghaastore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Ghaa.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
