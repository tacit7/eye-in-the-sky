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
end
