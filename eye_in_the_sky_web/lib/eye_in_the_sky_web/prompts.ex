defmodule EyeInTheSkyWeb.Prompts do
  @moduledoc """
  The Prompts context for managing subagent prompt templates.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Prompts.Prompt

  @doc """
  Returns the list of prompts with optional project filter.
  """
  def list_prompts(opts \\ []) do
    project_id = Keyword.get(opts, :project_id)
    include_inactive = Keyword.get(opts, :include_inactive, false)

    query = from(p in Prompt)

    query =
      if include_inactive do
        query
      else
        where(query, [p], p.active == true)
      end

    query =
      case project_id do
        nil -> query
        id -> where(query, [p], p.project_id == ^id or is_nil(p.project_id))
      end

    query
    |> order_by([p], desc: p.updated_at)
    |> Repo.all()
  end

  @doc """
  Gets a single prompt by ID.

  Raises `Ecto.NoResultsError` if the Prompt does not exist.
  """
  def get_prompt!(id), do: Repo.get!(Prompt, id)

  @doc """
  Gets a single prompt by slug.
  Returns nil if not found.
  """
  def get_prompt_by_slug(slug, project_id \\ nil) do
    query =
      from p in Prompt,
        where: p.slug == ^slug and p.active == true

    query =
      case project_id do
        nil -> where(query, [p], is_nil(p.project_id))
        id ->
          from p in query,
            where: p.project_id == ^id or is_nil(p.project_id),
            order_by: [desc: fragment("CASE WHEN ? IS NULL THEN 0 ELSE 1 END", p.project_id)]
      end

    Repo.one(query)
  end

  @doc """
  Creates a prompt.
  """
  def create_prompt(attrs \\ %{}) do
    %Prompt{}
    |> Prompt.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a prompt.
  """
  def update_prompt(%Prompt{} = prompt, attrs) do
    prompt
    |> Prompt.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Soft deletes a prompt by setting active to false.
  """
  def deactivate_prompt(%Prompt{} = prompt) do
    update_prompt(prompt, %{active: false})
  end

  @doc """
  Hard deletes a prompt.
  """
  def delete_prompt(%Prompt{} = prompt) do
    Repo.delete(prompt)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking prompt changes.
  """
  def change_prompt(%Prompt{} = prompt, attrs \\ %{}) do
    Prompt.changeset(prompt, attrs)
  end

  @doc """
  Lists global prompts (no project_id).
  """
  def list_global_prompts do
    Prompt
    |> where([p], is_nil(p.project_id) and p.active == true)
    |> order_by([p], desc: p.updated_at)
    |> Repo.all()
  end

  @doc """
  Lists project-specific prompts.
  """
  def list_project_prompts(project_id) do
    Prompt
    |> where([p], p.project_id == ^project_id and p.active == true)
    |> order_by([p], desc: p.updated_at)
    |> Repo.all()
  end

  @doc """
  Search prompts using FTS5.
  Requires prompt_search FTS5 table in database.
  """
  def search_prompts(query, project_id \\ nil) when is_binary(query) do
    sql = """
    SELECT p.*
    FROM subagent_prompts p
    JOIN prompt_search ps ON p.id = ps.rowid
    WHERE ps.prompt_search MATCH ?
    #{if project_id, do: "AND (p.project_id = ? OR p.project_id IS NULL)", else: ""}
    AND p.active = 1
    ORDER BY ps.rank
    LIMIT 50
    """

    params = if project_id, do: [query, project_id], else: [query]

    case Ecto.Adapters.SQL.query(Repo, sql, params) do
      {:ok, %{rows: rows, columns: columns}} ->
        Enum.map(rows, fn row ->
          columns
          |> Enum.zip(row)
          |> Map.new()
          |> then(&Repo.load(Prompt, &1))
        end)

      {:error, _} ->
        # Fallback to LIKE search if FTS5 table doesn't exist
        pattern = "%#{query}%"
        query_filter = from p in Prompt,
          where: (ilike(p.name, ^pattern) or ilike(p.description, ^pattern) or ilike(p.prompt_text, ^pattern)) and p.active == true

        query_filter = if project_id do
          where(query_filter, [p], p.project_id == ^project_id or is_nil(p.project_id))
        else
          query_filter
        end

        query_filter
        |> order_by([p], desc: p.updated_at)
        |> limit(50)
        |> Repo.all()
    end
  end
end
