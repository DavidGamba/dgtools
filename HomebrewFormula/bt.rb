class Bt < Formula
  @@tool_name = "bt"
  @@tool_desc = "A no commitments Terraform wrapper that provides build caching functionality"
  @@tool_path = "bt"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/bt/v0.11.0.tar.gz"
  sha256 "0c8613fa6055640840e3944dac60f1550e2c83cd43cb489da7f87a2295895abb"

  depends_on "go" => :build

  def install
    cd "#{@@tool_path}" do
      system "go", "get"
      system "go", "build"
      bin.install "#{@@tool_name}"
    end
    cd "HomebrewFormula" do
      inreplace "completions.bash", "tool", "#{@@tool_name}"
      inreplace "completions.zsh", "tool", "#{@@tool_name}"
      ohai "Installing bash completion..."
      bash_completion.install "completions.bash" => "dgtools.#{@@tool_name}.bash"
      ohai %{Installing zsh completion...
      To enable zsh completion add this to your ~/.zshrc

      \tsource #{zsh_completion.sub prefix, HOMEBREW_PREFIX}/dgtools.#{@@tool_name}.zsh
      }
      zsh_completion.install "completions.zsh" => "dgtools.#{@@tool_name}.zsh"
      ohai "Installed #{@@tool_name} from #{@@tool_path} dir"
    end
  end

  test do
    assert_match /Use '#{@@tool_name} help[^']*' for extra details/, shell_output("#{bin}/#{@@tool_name} --help")
  end
end
