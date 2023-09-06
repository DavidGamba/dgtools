class Tool
  attr_accessor :name, :desc, :path

  def initialize(name, desc, path)
    @name = name
    @desc = desc
    @path = path
  end
end
