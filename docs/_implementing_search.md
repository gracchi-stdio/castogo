An example of implementing a searchbox

```golang
			<form action="/search" method="GET" data-public-search class="hidden md:flex items-center">
				<div class="flex border-2 border-border">
					<input
						type="text"
						name="q"
						placeholder="SEARCH..."
						class="public-search-input border-none w-48"
						autocomplete="off"
					/>
					<button type="submit" class="px-3 text-muted-foreground hover:text-foreground transition-colors cursor-pointer">
						@icon.Lucide("search", "size-4")
					</button>
				</div>
			</form>
```
