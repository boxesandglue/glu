# Get version from git tag (e.g., "v1.0.0" or "v1.0.0-3-g1a2b3c4")
def git_version
  version = `git describe --tags --always --match 'v*' 2>/dev/null`.strip
  version.empty? ? "dev" : version.sub(/^v/, "")
end

@glu_version = git_version

desc "Show rake description"
task :default do
    puts
    puts "Run 'rake -T' for a list of tasks."
    puts
    puts "Use 'rake build' to build the 'glu' binary."
    puts
end

desc "Build the 'glu' binary"
task :build do
    sh "go build -ldflags '-s -w -X main.Version=#{@glu_version}' -o bin/glu github.com/boxesandglue/glu/glu"
end

desc "Install 'glu' into $GOBIN"
task :install do
    sh "go install -ldflags '-s -w -X main.Version=#{@glu_version}' github.com/boxesandglue/glu/glu"
end

desc "Show version information"
task :showversion do
    puts "glu version #{@glu_version}"
end

desc "Build the manpage (requires scdoc on PATH)"
task :manpage do
    src = "docs/glu.1.scd"
    out = "bin/glu.1"
    unless system("command -v scdoc >/dev/null 2>&1")
        abort "scdoc not found on PATH. Install via 'brew install scdoc' or your package manager."
    end
    FileUtils.mkdir_p("bin")
    sh "scdoc < #{src} > #{out}"
    puts "Wrote #{out}"
end

