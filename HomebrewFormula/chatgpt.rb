class Chatgpt < Formula
  @@tool_name = "chatgpt"
  @@tool_desc = "An OpenAI ChatGPT wrapper"
  @@tool_path = "chatgpt"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/chatgpt/v0.1.0.tar.gz"
  sha256 "de08fda70d2b2366d1dbd144cf8b1579c5fc7d0b6bdab32e67d1c24e49e518fd"

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
