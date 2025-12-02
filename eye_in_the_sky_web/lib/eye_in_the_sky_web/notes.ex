defmodule EyeInTheSkyWeb.Notes do
  @moduledoc """
  The Notes context for managing notes.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Notes.Note

  @doc """
  Returns the list of notes.
  """
  def list_notes do
    Repo.all(Note)
  end

  @doc """
  Returns notes for a specific session.
  """
  def list_notes_for_session(session_id) do
    Note
    |> where([n], n.parent_type == "session" and n.parent_id == ^session_id)
    |> order_by([n], desc: n.created_at)
    |> Repo.all()
  end

  @doc """
  Counts notes for a specific session.
  """
  def count_notes_for_session(session_id) do
    Note
    |> where([n], n.parent_type == "session" and n.parent_id == ^session_id)
    |> Repo.aggregate(:count, :id)
  end

  @doc """
  Returns notes for a specific agent.
  """
  def list_notes_for_agent(agent_id) do
    Note
    |> where([n], n.parent_type == "agent" and n.parent_id == ^agent_id)
    |> order_by([n], desc: n.created_at)
    |> Repo.all()
  end

  @doc """
  Gets a single note.
  """
  def get_note!(id) do
    Repo.get!(Note, id)
  end

  @doc """
  Creates a note.
  """
  def create_note(attrs \\ %{}) do
    %Note{}
    |> Note.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Deletes a note.
  """
  def delete_note(%Note{} = note) do
    Repo.delete(note)
  end

  @doc """
  Search notes using FTS5.
  Requires note_search FTS5 table in database.
  """
  def search_notes(query, agent_ids \\ []) when is_binary(query) do
    sql = """
    SELECT n.*
    FROM notes n
    JOIN note_search ns ON n.id = ns.rowid
    WHERE ns.note_search MATCH ?
    #{if length(agent_ids) > 0, do: "AND n.parent_type = 'agent' AND n.parent_id IN (#{Enum.map(agent_ids, fn _ -> "?" end) |> Enum.join(",")})", else: ""}
    ORDER BY ns.rank
    LIMIT 50
    """

    params = if length(agent_ids) > 0, do: [query | agent_ids], else: [query]

    case Ecto.Adapters.SQL.query(Repo, sql, params) do
      {:ok, %{rows: rows, columns: columns}} ->
        Enum.map(rows, fn row ->
          columns
          |> Enum.zip(row)
          |> Map.new()
          |> then(&Repo.load(Note, &1))
        end)

      {:error, _} ->
        # Fallback to LIKE search if FTS5 table doesn't exist
        pattern = "%#{query}%"
        query_filter = from n in Note,
          where: ilike(n.body, ^pattern)

        query_filter = if length(agent_ids) > 0 do
          where(query_filter, [n], n.parent_type == "agent" and n.parent_id in ^agent_ids)
        else
          query_filter
        end

        query_filter
        |> order_by([n], desc: n.created_at)
        |> limit(50)
        |> Repo.all()
    end
  end
end
