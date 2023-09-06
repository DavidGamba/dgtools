load './tool.rb'

class Bt < Formula
  @@tool = Tool.new("bt", "A no commitments Terraform wrapper that provides build caching functionality", "bt")

  desc "#{@@tool.desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool.name}"
  head "https://github.com/DavidGamba/dgtools.git", branch: "master"

  depends_on "go" => :build

  def install
    cd "#{@@tool.path}" do
      system "go", "get"
      system "go", "build"
      bin.install "#{@@tool.name}"
    end
    cd "HomebrewFormula" do
      inreplace "completions.bash", "tool", "#{@@tool.name}"
      inreplace "completions.zsh", "tool", "#{@@tool.name}"
      ohai "Installing bash completion..."
      bash_completion.install "completions.bash" => "dgtools.#{@@tool.name}.bash"
      ohai %{Installing zsh completion...
      To enable zsh completion add this to your ~/.zshrc

      \tsource #{zsh_completion.sub prefix, HOMEBREW_PREFIX}/dgtools.#{@@tool.name}.zsh
      }
      zsh_completion.install "completions.zsh" => "dgtools.#{@@tool.name}.zsh"
      ohai "Installed #{@@tool.name} from #{@@tool.path} dir"
    end
  end

  test do
    assert_match /Use '#{@@tool.name} help[^']*' for extra details/, shell_output("#{bin}/#{@@tool.name} --help")
  end
end
