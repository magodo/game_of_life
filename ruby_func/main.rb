-> (x_dimension, y_dimension, n_day, day_interval_sec: 0.05) do
  cur_grid = Array.new(x_dimension) {Array.new(y_dimension) {rand >= 0.5}}
  n_day.times do
    start = Time.now
    cur_grid = -> (grid, transit) do
      grid.each_with_index.map do |row, x|
        row.each_with_index.map do |state, y|
          neighbours = [x - 1, x, x + 1].product([y - 1, y, y + 1])
                           .reject { |_x, _y| (_x == x && _y == y) }
                           .select { |_x, _y| (0...grid.size).include?(_x) && (0...grid.first.size).include?(_y) }
                           .map { |_x, _y| grid[_x][_y] }
          transit.call(state, neighbours)
        end
      end
    end.call(
        cur_grid,
        -> (state, neighbours) {
          {
              true: ->() { (2..3).include? neighbours.select { |is_alive| is_alive == true }.size },
              false: ->() { 3 == neighbours.select { |is_alive| is_alive == true }.size },
          }[state.to_s.to_sym].call
        },
    ).tap {| _grid| system 'clear'; _grid.map { |row| puts row.map { |state| state ? '*' : ' ' }.join(' ') } }
    puts "#{Time.now - start} sec"
    sleep(day_interval_sec) 
  end
end.call(
       150,
       200,
    999999,
)
