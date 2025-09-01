function populateExerciseTable(exercises) {
    const tbody = document.querySelector(".exercise-table tbody");
    tbody.innerHTML = "";

    Object.values(exercises).forEach(ex => {
        const row = document.createElement("tr");
        row.innerHTML = `
            <td>${ex.ex_name}</td>
            <td>${ex.sets}</td>
            <td>${ex.reps}</td>
            <td>${ex.weight}</td>
            <td>${ex.notes || ""}</td>
        `;
        tbody.appendChild(row);
    });
}

document.addEventListener("DOMContentLoaded", () => {
    fetch("/allworkouts")
        .then(response => response.json())
        .then(res => {
            if (!res.success) {
                console.error("API error:", res.error);
                return;
            }
            const workouts = res.data; // this is your array

            const workoutTable = document.querySelector(".workouts-table").getElementsByTagName("tbody")[0];
            workoutTable.innerHTML = ""; // clear existing

            workouts.forEach(workout => {
                const row = workoutTable.insertRow();

                // Date
                const dateCell = row.insertCell(0);
                const date = new Date(workout.date);
                dateCell.textContent = date.toISOString().split("T")[0]; // → "2025-07-12"

                // Type
                const typeCell = row.insertCell(1);
                typeCell.textContent = workout.wtype;

                // View button
                const viewCell = row.insertCell(2);
                const viewBtn = document.createElement("button");
                viewBtn.textContent = "View";
                viewBtn.classList.add("view-btn");
                viewCell.appendChild(viewBtn);

                viewBtn.addEventListener("click", () => {
                    let exercises;
                    try {
                        exercises = JSON.parse(workout.data);
                    } catch (e) {
                        console.error("Failed to parse exercise data:", workout.data, e);
                        return;
                    }

                    populateExerciseTable(exercises)

                    const modal = document.getElementById("workoutModal");
                    modal.dataset.currentDate = workout.date.split("T")[0]; // store clean date
                    modal.style.display = "flex";
                });

                // Delete button
                const deleteBtn = document.createElement("button");
                deleteBtn.textContent = "Delete";
                deleteBtn.classList.add("delete-btn");
                viewCell.appendChild(deleteBtn);

                deleteBtn.addEventListener("click", () => {
                    if (!confirm(`Delete workout on ${workout.date}?`)) return;

                    const dateOnly = workout.date.split("T")[0];

                    fetch(`/removeworkout/${encodeURIComponent(dateOnly)}`, {
                        method: "DELETE"
                    })
                        .then(response => response.json())
                        .then(res => {
                            if (!res.success) {
                                console.error("Delete failed:", res.error);
                                return;
                            }
                            // remove row from UI
                            row.remove();
                        })
                        .catch(err => console.error("Failed to delete workout:", err));
                });
            });
        })
        .catch(err => console.error("Failed to load workouts:", err));
});

document.addEventListener("DOMContentLoaded", () => {
    const modal = document.getElementById("workoutModal");
    const closeModalBtn = modal.querySelector(".close-modal-btn");
    const viewButtons = document.querySelectorAll(".view-btn");
    const addExerciseBtn = modal.querySelector(".add-exercise-btn");
    const deleteExerciseBtn = modal.querySelector(".delete-exercise-btn");
    const saveExercisesBtn = modal.querySelector(".save-exercise-btn")
    const exerciseTable = modal.querySelector(".exercise-table tbody");
    const addWorkoutBtn = document.querySelector(".add-workout-btn");
    const mainSect = document.getElementById("main-section");
    const workoutTable = mainSect.querySelector(".workouts-table tbody")

    addWorkoutBtn.addEventListener("click", async() => {
        const date = prompt("Enter workout date (YYYY-MM-DD):");

        if (!date.match(/^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$/)) {
            alert("The date you entered is not valid. Please try again:");
            return;
        }

        const type = prompt("Enter workout type (Push, Pull, Legs, etc):");

        if (!date || !type) {
            alert("Please provide both date and type.");
            return;
        }

        const workout = {
            date: date,
            wtype: type,
            data: "{}" // empty JSON string for now
        };

        try {
            const response = await fetch("/addworkout", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(workout)
            });

            const result = await response.json();

            if (response.ok) {
                // Add a row to the table
                const row = document.createElement("tr");
                row.innerHTML = `
                <td contenteditable="true">${date}</td>
                <td contenteditable="true">${type}</td>
                <td>
                    <div class="action-btns">
                        <button class="view-btn">View</button>
                        <button class="delete-btn">Delete</button>
                    </div>
                </td>
            `;
                workoutTable.appendChild(row);

                console.log("Workout added:", result);
            } else {
                alert("Error: " + result.error);
            }
        } catch (err) {
            console.error("Failed to add workout:", err);
        }
    });

    viewButtons.forEach(btn => {
        btn.addEventListener("click", () => {
            modal.style.display = "flex";
        });
    });

    // Open modal when clicking "View"
    workoutTable.addEventListener("click", (event) => {
        if (event.target.classList.contains("view-btn")) {
            modal.style.display = "flex";
        }
    });

    // Close modal
    closeModalBtn.addEventListener("click", () => {
        modal.style.display = "none";
    });

    // Add new exercise row
    addExerciseBtn.addEventListener("click", () => {
        const newRow = document.createElement("tr");
        newRow.innerHTML = `
            <td contenteditable="true">New Exercise</td>
            <td contenteditable="true">0</td>
            <td contenteditable="true">0</td>
            <td contenteditable="true">0kg</td>
            <td contenteditable="true">Notes...</td>
          `;
        exerciseTable.appendChild(newRow);
    });

    // Delete last exercise row
    deleteExerciseBtn.addEventListener("click", () => {
        if (exerciseTable.rows.length > 0) {
            exerciseTable.deleteRow(exerciseTable.rows.length - 1);
        }
    });

    saveExercisesBtn.addEventListener("click", () => {
        const rows = exerciseTable.querySelectorAll("tr");
        const exercises = {};

        rows.forEach((row, i) => {
            const cells = row.querySelectorAll("td");
            exercises[`exercise_${i}`] = {
                ex_name: cells[0].textContent.trim(),
                sets: parseInt(cells[1].textContent.trim(), 10) || 0,
                reps: parseInt(cells[2].textContent.trim(), 10) || 0,
                weight: parseFloat(cells[3].textContent.trim()) || 0,
                notes: cells[4].textContent.trim()
            };
        });

        const workoutDate = modal.dataset.currentDate;

        fetch(`/updateworkout/${workoutDate}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ data: JSON.stringify(exercises) })
        })
            .then(res => res.json())
            .then(result => {
                if (result.success) {
                    alert("Workout updated successfully!");
                    const newExercises = JSON.parse(result.data.data);
                    populateExerciseTable(newExercises);
                    location.reload();
                } else {
                    alert("Error updating workout: " + result.error);
                }
            })
            .catch(err => {
                console.error("Update failed:", err);
                alert("An error occurred while saving.");
            });
    });

    // Close modal if clicking outside
    window.addEventListener("click", (event) => {
        if (event.target === modal) {
            modal.style.display = "none";
        }
    });
});