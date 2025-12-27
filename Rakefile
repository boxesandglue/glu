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
    sh "go build -ldflags '-s -w -X main.Version=#{@glu_version}' -o bin/glu github.com/speedata/glu/glu"
end

desc "Show version information"
task :showversion do
    puts "glu version #{@glu_version}"
end

desc "Clean build artifacts"
task :clean do
    FileUtils.rm_rf("bin")
end
